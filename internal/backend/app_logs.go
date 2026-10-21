package backend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"ostenia/internal/config"
	"ostenia/internal/service"
)

var (
	logMutex        sync.Mutex
	currentLogDir   string
	currentLogFile  string
	currentLogDate  string
	currentLogLines int
	currentLogSeq   int
)

// CallerInfo represents details of the caller function for the log entry.
type CallerInfo struct {
	FunctionName string `json:"functionName,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	Line         string `json:"line,omitempty"`
	Column       string `json:"column,omitempty"`
}

// LogEntry represents a structured log entry stored in JSONL and displayed in the frontend.
type LogEntry struct {
	ID           string      `json:"id"`
	Time         string      `json:"time"`
	Timestamp    string      `json:"timestamp,omitempty"`
	Msg          string      `json:"msg"`
	CleanMsg     string      `json:"cleanMsg,omitempty"`
	Type         string      `json:"type,omitempty"` // "info", "warn", "error"
	IsServiceLog bool        `json:"isServiceLog"`
	Caller       *CallerInfo `json:"caller,omitempty"`
	RawStack     string      `json:"rawStack,omitempty"`
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

// findLatestLogFile scans the logs directory for today's latest JSONL log file and returns its path, sequence number, and line count.
func findLatestLogFile(logsDir, dateStr string) (string, int, int, error) {
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return "", 0, 0, err
	}

	maxSeq := 0
	var latestFile string
	prefix := fmt.Sprintf("ostenia-%s-", dateStr)

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".jsonl") {
			seqStr := name[len(prefix) : len(name)-len(".jsonl")]
			var seq int
			_, err := fmt.Sscanf(seqStr, "%d", &seq)
			if err == nil && seq > maxSeq {
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

// pruneOldLogs removes log files older than the specified retention days.
func pruneOldLogs(logsDir string, retentionDays int) {
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, entry := range files {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(logsDir, entry.Name()))
		}
	}
}

// SaveLogEntry appends a structured log entry as a JSON line in the persistent local AppData directory.
func (a *App) SaveLogEntry(entry LogEntry) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	logsDir := config.GetLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	now := time.Now()
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%d", now.UnixNano())
	}
	if entry.Time == "" {
		entry.Time = now.Format("15:04:05")
	}
	if entry.Timestamp == "" {
		entry.Timestamp = now.Format(time.RFC3339)
	}
	if entry.Type == "" {
		entry.Type = "info"
	}
	if entry.CleanMsg == "" {
		entry.CleanMsg = entry.Msg
	}

	dateStr := now.Format("2006-01-02")
	if logsDir != currentLogDir || dateStr != currentLogDate || currentLogFile == "" {
		currentLogDir = logsDir
		currentLogDate = dateStr
		latestFile, maxSeq, lines, err := findLatestLogFile(logsDir, dateStr)
		if err != nil {
			return err
		}
		currentLogFile = latestFile
		currentLogSeq = maxSeq
		currentLogLines = lines
	}

	targetFile := currentLogFile
	if currentLogFile == "" || currentLogLines >= 1000 {
		currentLogSeq++
		fileName := fmt.Sprintf("ostenia-%s-%02d.jsonl", dateStr, currentLogSeq)
		currentLogFile = filepath.Join(logsDir, fileName)
		currentLogLines = 0
		targetFile = currentLogFile
		go pruneOldLogs(logsDir, 14)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(append(data, '\n')); err != nil {
		return err
	}

	currentLogLines++
	return nil
}

// SaveLogToFile maintains backward compatibility, accepting a string message and saving it as a structured log entry.
func (a *App) SaveLogToFile(message string) error {
	entry := LogEntry{
		Msg:      message,
		CleanMsg: message,
		Type:     "info",
	}

	if strings.HasPrefix(message, "[ERR]") {
		entry.Type = "error"
		entry.CleanMsg = strings.TrimSpace(strings.TrimPrefix(message, "[ERR]"))
	} else if strings.HasPrefix(message, "[WRN]") {
		entry.Type = "warn"
		entry.CleanMsg = strings.TrimSpace(strings.TrimPrefix(message, "[WRN]"))
	} else if strings.HasPrefix(message, "[SYS]") {
		entry.CleanMsg = strings.TrimSpace(strings.TrimPrefix(message, "[SYS]"))
	}

	return a.SaveLogEntry(entry)
}

// GetRecentLogs retrieves the most recent log entries across JSONL log files up to the specified limit.
// Returned entries are ordered newest first (index 0 is newest).
func (a *App) GetRecentLogs(limit int) ([]LogEntry, error) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if limit <= 0 {
		limit = 200
	} else if limit > 2000 {
		limit = 2000
	}

	logsDir := config.GetLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, err
	}

	var jsonlFiles []string
	for _, entry := range files {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "ostenia-") && strings.HasSuffix(entry.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, filepath.Join(logsDir, entry.Name()))
		}
	}

	if len(jsonlFiles) == 0 {
		return []LogEntry{}, nil
	}

	// Sort files newest first by filename (e.g. ostenia-2026-10-21-02.jsonl after 01)
	sort.Sort(sort.Reverse(sort.StringSlice(jsonlFiles)))

	var result []LogEntry
	for _, filePath := range jsonlFiles {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		var fileEntries []LogEntry
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var entry LogEntry
			if err := json.Unmarshal(line, &entry); err == nil {
				fileEntries = append(fileEntries, entry)
			}
		}
		file.Close()

		// fileEntries are in chronological order (oldest -> newest).
		// Reverse them to get newest first.
		for i := len(fileEntries) - 1; i >= 0; i-- {
			result = append(result, fileEntries[i])
			if len(result) >= limit {
				return result, nil
			}
		}
	}

	return result, nil
}

// OpenLogsFolder opens the local AppData logs directory in the operating system's file explorer.
func (a *App) OpenLogsFolder() error {
	logsDir := config.GetLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}
	return service.OpenExplorer(logsDir)
}

// ClearLogs deletes all persistent JSONL log files in the logs directory.
func (a *App) ClearLogs() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	logsDir := config.GetLogsDir()
	files, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			currentLogFile = ""
			currentLogLines = 0
			currentLogSeq = 0
			return nil
		}
		return err
	}

	for _, entry := range files {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".jsonl") || strings.HasSuffix(entry.Name(), ".log")) {
			_ = os.Remove(filepath.Join(logsDir, entry.Name()))
		}
	}

	currentLogFile = ""
	currentLogLines = 0
	currentLogSeq = 0
	return nil
}
