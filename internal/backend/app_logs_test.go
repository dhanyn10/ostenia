package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func verifyJsonlFileLines(t *testing.T, filePath string, expectedLines int) {
	t.Helper()
	lines, err := countLines(filePath)
	if err != nil {
		t.Fatalf("Failed to count lines in %s: %v", filePath, err)
	}
	if lines != expectedLines {
		t.Errorf("Expected file %s line count to be %d, got %d", filePath, expectedLines, lines)
	}
}

func TestGetLogsDir(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Test override
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	if GetLogsDir() != tempDir {
		t.Fatalf("Expected override dir %s, got %s", tempDir, GetLogsDir())
	}

	SetLogDirOverride("")

	// 2. Test OSTENIA_LOGS_DIR
	envDir := filepath.Join(tempDir, "custom-logs")
	oldEnv := os.Getenv("OSTENIA_LOGS_DIR")
	os.Setenv("OSTENIA_LOGS_DIR", envDir)
	defer os.Setenv("OSTENIA_LOGS_DIR", oldEnv)

	if GetLogsDir() != envDir {
		t.Fatalf("Expected env dir %s, got %s", envDir, GetLogsDir())
	}
	os.Unsetenv("OSTENIA_LOGS_DIR")

	// 3. Test LOCALAPPDATA
	oldLocal := os.Getenv("LOCALAPPDATA")
	fakeLocalAppData := filepath.Join(tempDir, "AppData", "Local")
	os.Setenv("LOCALAPPDATA", fakeLocalAppData)
	defer os.Setenv("LOCALAPPDATA", oldLocal)

	expected := filepath.Join(fakeLocalAppData, "Ostenia", "logs")
	if GetLogsDir() != expected {
		t.Fatalf("Expected LOCALAPPDATA dir %s, got %s", expected, GetLogsDir())
	}
}

func TestApp_SaveLogEntry_And_Rotation(t *testing.T) {
	tempDir := t.TempDir()
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	app := NewApp()
	dateStr := time.Now().Format("2006-01-02")
	expectedFile1 := filepath.Join(tempDir, fmt.Sprintf("ostenia-%s-01.jsonl", dateStr))

	// 1. Write first structured log entry
	entry1 := LogEntry{
		ID:           "test-id-1",
		Time:         "10:00:00",
		Msg:          "[SYS] Service starting",
		CleanMsg:     "Service starting",
		Type:         "info",
		IsServiceLog: false,
		Caller: &CallerInfo{
			FunctionName: "startService",
			FileName:     "main.go",
			Line:         "42",
		},
	}

	if err := app.SaveLogEntry(entry1); err != nil {
		t.Fatalf("SaveLogEntry failed: %v", err)
	}

	content, err := os.ReadFile(expectedFile1)
	if err != nil {
		t.Fatalf("Expected file to exist at %s, error: %v", expectedFile1, err)
	}

	var parsed LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &parsed); err != nil {
		t.Fatalf("Failed to parse JSONL line: %v", err)
	}
	if parsed.ID != "test-id-1" || parsed.Msg != "[SYS] Service starting" || parsed.Caller == nil || parsed.Caller.Line != "42" {
		t.Errorf("Unexpected parsed entry: %+v", parsed)
	}

	// 2. Write 999 more entries to reach maxLinesPerFile (1000)
	for i := 2; i <= 1000; i++ {
		err := app.SaveLogEntry(LogEntry{
			ID:       fmt.Sprintf("test-id-%d", i),
			Msg:      fmt.Sprintf("Log message %d", i),
			CleanMsg: fmt.Sprintf("Log message %d", i),
			Type:     "info",
		})
		if err != nil {
			t.Fatalf("Failed to write line %d: %v", i, err)
		}
	}
	verifyJsonlFileLines(t, expectedFile1, 1000)

	// 3. Write 1001st entry, should rotate to sequence 02
	entry1001 := LogEntry{
		ID:       "test-id-1001",
		Msg:      "Rotated line",
		CleanMsg: "Rotated line",
		Type:     "warn",
	}
	if err := app.SaveLogEntry(entry1001); err != nil {
		t.Fatalf("Failed to write rotating line: %v", err)
	}

	expectedFile2 := filepath.Join(tempDir, fmt.Sprintf("ostenia-%s-02.jsonl", dateStr))
	verifyJsonlFileLines(t, expectedFile2, 1)
	verifyJsonlFileLines(t, expectedFile1, 1000)
}

func TestApp_SaveLogToFile_Compatibility(t *testing.T) {
	tempDir := t.TempDir()
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	app := NewApp()
	msg := "[WRN] Database connection timeout"
	if err := app.SaveLogToFile(msg); err != nil {
		t.Fatalf("SaveLogToFile failed: %v", err)
	}

	dateStr := time.Now().Format("2006-01-02")
	expectedFile := filepath.Join(tempDir, fmt.Sprintf("ostenia-%s-01.jsonl", dateStr))

	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("File not created: %v", err)
	}

	var parsed LogEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(content))), &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if parsed.Msg != msg || parsed.CleanMsg != msg {
		t.Errorf("Unexpected parsed message: %s", parsed.Msg)
	}
}

func TestApp_GetRecentLogs(t *testing.T) {
	tempDir := t.TempDir()
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	app := NewApp()

	// Empty dir returns empty slice
	emptyLogs, err := app.GetRecentLogs(10)
	if err != nil {
		t.Fatalf("Expected nil error for empty dir, got: %v", err)
	}
	if len(emptyLogs) != 0 {
		t.Fatalf("Expected 0 logs, got %d", len(emptyLogs))
	}

	// Write 5 entries
	for i := 1; i <= 5; i++ {
		_ = app.SaveLogEntry(LogEntry{
			ID:       fmt.Sprintf("log-%d", i),
			Msg:      fmt.Sprintf("Message %d", i),
			CleanMsg: fmt.Sprintf("Message %d", i),
		})
	}

	// Read with limit 3 (should return newest first: log-5, log-4, log-3)
	logs, err := app.GetRecentLogs(3)
	if err != nil {
		t.Fatalf("GetRecentLogs failed: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("Expected 3 logs, got %d", len(logs))
	}
	if logs[0].ID != "log-5" || logs[1].ID != "log-4" || logs[2].ID != "log-3" {
		t.Errorf("Unexpected log order: %v, %v, %v", logs[0].ID, logs[1].ID, logs[2].ID)
	}

	// Read with limit 0 (defaults to 500)
	allLogs, err := app.GetRecentLogs(0)
	if err != nil {
		t.Fatalf("GetRecentLogs(0) failed: %v", err)
	}
	if len(allLogs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(allLogs))
	}
}

func TestApp_ClearLogs(t *testing.T) {
	tempDir := t.TempDir()
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	app := NewApp()
	_ = app.SaveLogToFile("First entry")
	_ = app.SaveLogToFile("Second entry")

	logs, _ := app.GetRecentLogs(10)
	if len(logs) == 0 {
		t.Fatalf("Logs should exist before clearing")
	}

	if err := app.ClearLogs(); err != nil {
		t.Fatalf("ClearLogs failed: %v", err)
	}

	logsAfter, _ := app.GetRecentLogs(10)
	if len(logsAfter) != 0 {
		t.Errorf("Expected 0 logs after clearing, got %d", len(logsAfter))
	}
}

func TestApp_CleanOldLogs(t *testing.T) {
	tempDir := t.TempDir()
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	// Create an ancient log file (> 15 days old)
	oldFilePath := filepath.Join(tempDir, "ostenia-2020-01-01-01.jsonl")
	if err := os.WriteFile(oldFilePath, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}
	oldTime := time.Now().AddDate(0, 0, -20)
	_ = os.Chtimes(oldFilePath, oldTime, oldTime)

	// Create a fresh log file
	freshFilePath := filepath.Join(tempDir, "ostenia-fresh-01.jsonl")
	if err := os.WriteFile(freshFilePath, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("Failed to create fresh file: %v", err)
	}

	cleanOldLogs(tempDir)

	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Errorf("Expected ancient log file to be deleted, but it exists")
	}
	if _, err := os.Stat(freshFilePath); os.IsNotExist(err) {
		t.Errorf("Expected fresh log file to remain, but it was deleted")
	}
}

func TestApp_OpenLogsFolder(t *testing.T) {
	tempDir := t.TempDir()
	oldOverride := SetLogDirOverride(tempDir)
	defer SetLogDirOverride(oldOverride)

	app := NewApp()
	// OpenLogsFolder creates the dir and invokes OpenExplorer
	// In test environment, OpenExplorer might fail to launch GUI, but MkdirAll must succeed
	_ = app.OpenLogsFolder()

	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Expected logs folder to exist")
	}
}
