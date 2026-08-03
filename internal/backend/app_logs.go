package backend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"ostenia/internal/config"
)

var logMutex sync.Mutex

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

// findLatestLogFile scans the base directory for today's latest log file and returns its path, sequence number, and line count.
func findLatestLogFile(baseDir, dateStr string) (string, int, int, error) {
	files, err := os.ReadDir(baseDir)
	if err != nil {
		return "", 0, 0, err
	}

	maxSeq := 0
	var latestFile string

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match ddmmyy-xx.log (length 13, e.g., 030825-01.log)
		if len(name) == 13 && name[:6] == dateStr && name[6] == '-' && name[9:] == ".log" {
			seqStr := name[7:9]
			var seq int
			_, err := fmt.Sscanf(seqStr, "%d", &seq)
			if err == nil && seq > maxSeq {
				maxSeq = seq
				latestFile = filepath.Join(baseDir, name)
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

// SaveLogToFile appends a log message to the current log file, rotating to a new file after 1000 lines.
func (a *App) SaveLogToFile(message string) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	baseDir := config.GetBaseDir()
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}

	dateStr := time.Now().Format("020106")
	latestFile, maxSeq, lines, err := findLatestLogFile(baseDir, dateStr)
	if err != nil {
		return err
	}

	targetFile := latestFile
	if latestFile == "" || lines >= 1000 {
		nextSeq := maxSeq + 1
		fileName := fmt.Sprintf("%s-%02d.log", dateStr, nextSeq)
		targetFile = filepath.Join(baseDir, fileName)
	}

	file, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)
	_, err = file.WriteString(logLine)
	return err
}
