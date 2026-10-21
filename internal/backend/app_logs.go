package backend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ostenia/internal/config"
	"ostenia/internal/service"
)

var (
	logMutex        sync.Mutex
	logDirOverride  string
	maxLinesPerFile = 1000

	activeLogFile  string
	activeLogLines int
	activeLogDate  string
	activeLogSeq   int
	lastCleanTime  time.Time
)

const logRetentionDays = 14

// CallerInfo details where a frontend or backend log was triggered.
type CallerInfo struct {
	FunctionName string `json:"functionName,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	Line         string `json:"line,omitempty"`
	Column       string `json:"column,omitempty"`
}

// LogEntry represents an individual structured log event serialized to JSONL.
type LogEntry struct {
	ID           string      `json:"id"`
	Time         string      `json:"time"`
	Timestamp    string      `json:"timestamp,omitempty"`
	Msg          string      `json:"msg"`
	CleanMsg     string      `json:"cleanMsg,omitempty"`
	Type         string      `json:"type,omitempty"`
	IsServiceLog bool        `json:"isServiceLog"`
	Caller       *CallerInfo `json:"caller,omitempty"`
	RawStack     string      `json:"rawStack,omitempty"`
}

// SetLogDirOverride allows setting a custom logs directory for testing.
// It returns the previous override value.
func SetLogDirOverride(path string) string {
	logMutex.Lock()
	defer logMutex.Unlock()
	old := logDirOverride
	logDirOverride = path
	activeLogFile = ""
	activeLogLines = 0
	return old
}

// GetLogsDir returns the directory path for storing Ostenia logs in Local AppData.
// Priority:
// 1. logDirOverride (for tests)
// 2. OSTENIA_LOGS_DIR environment variable
// 3. %LOCALAPPDATA%\Ostenia\logs (Windows)
// 4. os.UserCacheDir() / os.UserConfigDir() (Cross-platform)
// 5. BaseDir/logs fallback
func GetLogsDir() string {
	if logDirOverride != "" {
		return logDirOverride
	}

	if env := os.Getenv("OSTENIA_LOGS_DIR"); env != "" {
		return env
	}

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		return filepath.Join(localAppData, "Ostenia", "logs")
	}

	if cacheDir, err := os.UserCacheDir(); err == nil && cacheDir != "" {
		return filepath.Join(cacheDir, "Ostenia", "logs")
	}

	return filepath.Join(config.GetBaseDir(), "logs")
}

// countLines counts the number of lines in a file.
func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	return lines, scanner.Err()
}

// findLatestJsonlFile scans logsDir for today's latest JSONL log file and returns its path, sequence number, and line count.
func findLatestJsonlFile(logsDir, dateStr string) (string, int, int, error) {
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return "", 0, 0, err
	}

	prefix := fmt.Sprintf("ostenia-%s-", dateStr)
	suffix := ".jsonl"
	maxSeq := 0
	var latestFile string

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
			seqStr := name[len(prefix) : len(name)-len(suffix)]
			seq, parseErr := strconv.Atoi(seqStr)
			if parseErr == nil && seq > maxSeq {
				maxSeq = seq
				latestFile = filepath.Join(logsDir, name)
			}
		}
	}

	if latestFile == "" {
		return "", 0, 0, nil
	}

	lines, err := countLines(latestFile)
	if err != nil {
		return "", 0, 0, err
	}

	return latestFile, maxSeq, lines, nil
}

// cleanOldLogs removes log files older than logRetentionDays.
func cleanOldLogs(logsDir string) {
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -logRetentionDays)
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".jsonl") || strings.HasSuffix(name, ".log") {
			fullPath := filepath.Join(logsDir, name)
			info, err := entry.Info()
			if err == nil && info.ModTime().Before(cutoff) {
				_ = os.Remove(fullPath)
			}
		}
	}
}

// SaveLogEntry appends a structured log entry as a JSON line to today's log file in Local AppData.
func (a *App) SaveLogEntry(entry LogEntry) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	logsDir := GetLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}
	if entry.Time == "" {
		entry.Time = time.Now().Format("15:04:05")
	}

	dateStr := time.Now().Format("2006-01-02")

	if activeLogFile == "" || activeLogDate != dateStr || activeLogLines >= maxLinesPerFile {
		if time.Since(lastCleanTime) > 12*time.Hour {
			cleanOldLogs(logsDir)
			lastCleanTime = time.Now()
		}

		latestFile, maxSeq, lines, err := findLatestJsonlFile(logsDir, dateStr)
		if err != nil {
			return err
		}

		if latestFile == "" || lines >= maxLinesPerFile {
			activeLogSeq = maxSeq + 1
			activeLogFile = filepath.Join(logsDir, fmt.Sprintf("ostenia-%s-%02d.jsonl", dateStr, activeLogSeq))
			activeLogLines = 0
		} else {
			activeLogFile = latestFile
			activeLogSeq = maxSeq
			activeLogLines = lines
		}
		activeLogDate = dateStr
	}

	file, err := os.OpenFile(activeLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	if _, err = file.Write(data); err != nil {
		return err
	}

	activeLogLines++
	return nil
}

// SaveLogToFile formats a plain text message as a LogEntry and persists it via SaveLogEntry.
// Kept for backwards compatibility.
func (a *App) SaveLogToFile(message string) error {
	entry := LogEntry{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Time:         time.Now().Format("15:04:05"),
		Timestamp:    time.Now().Format(time.RFC3339),
		Msg:          message,
		CleanMsg:     message,
		Type:         "info",
		IsServiceLog: false,
	}
	return a.SaveLogEntry(entry)
}

// GetRecentLogs loads up to limit entries from JSONL log files in Local AppData, returned newest first.
func (a *App) GetRecentLogs(limit int) ([]LogEntry, error) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if limit <= 0 {
		limit = 500
	}

	logsDir := GetLogsDir()
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return []LogEntry{}, nil
	}

	files, err := os.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}

	var jsonlFiles []string
	for _, entry := range files {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, filepath.Join(logsDir, entry.Name()))
		}
	}
	sort.Strings(jsonlFiles)

	var allEntries []LogEntry

	// Read backwards from newest files
	for i := len(jsonlFiles) - 1; i >= 0; i-- {
		filePath := jsonlFiles[i]
		entries, err := readLogEntriesFromFile(filePath)
		if err != nil {
			continue
		}
		for j := len(entries) - 1; j >= 0; j-- {
			allEntries = append(allEntries, entries[j])
			if len(allEntries) >= limit {
				return allEntries, nil
			}
		}
	}

	if allEntries == nil {
		return []LogEntry{}, nil
	}
	return allEntries, nil
}

func readLogEntriesFromFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

// ClearLogs removes all log files in the Local AppData logs folder.
func (a *App) ClearLogs() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	activeLogFile = ""
	activeLogLines = 0

	logsDir := GetLogsDir()
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return nil
	}

	files, err := os.ReadDir(logsDir)
	if err != nil {
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".jsonl") || strings.HasSuffix(entry.Name(), ".log")) {
			_ = os.Remove(filepath.Join(logsDir, entry.Name()))
		}
	}
	return nil
}

// OpenLogsFolder opens the Local AppData logs directory in the default system file manager.
func (a *App) OpenLogsFolder() error {
	logsDir := GetLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}
	return service.OpenExplorer(logsDir)
}
