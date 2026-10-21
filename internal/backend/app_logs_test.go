package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ostenia/internal/service"
)

func verifyLogFileContentAndLines(t *testing.T, filePath, expectedMessage string, expectedLines int) {
	t.Helper()
	if expectedMessage != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Expected file to exist at %s, got error: %v", filePath, err)
		}
		if !strings.Contains(string(content), expectedMessage) {
			t.Errorf("Expected content to contain '%s', got: %s", expectedMessage, string(content))
		}
	}

	lines, err := countLines(filePath)
	if err != nil {
		t.Fatalf("Failed to count lines: %v", err)
	}
	if lines != expectedLines {
		t.Errorf("Expected file %s line count to be %d, got %d", filePath, expectedLines, lines)
	}
}

func writeLogLinesBatch(t *testing.T, app *App, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		if err := app.SaveLogToFile(fmt.Sprintf("Log line %d", i)); err != nil {
			t.Fatalf("Failed to write log line at index %d: %v", i, err)
		}
	}
}

func TestApp_SaveLogToFile_Rotation(t *testing.T) {
	tempDir := t.TempDir()

	// Direct logs to tempDir via OSTENIA_LOGS_DIR
	oldLogsDir := os.Getenv("OSTENIA_LOGS_DIR")
	os.Setenv("OSTENIA_LOGS_DIR", tempDir)
	defer os.Setenv("OSTENIA_LOGS_DIR", oldLogsDir)

	app := NewApp()

	// 1. Write first log message
	message1 := "First test log message"
	if err := app.SaveLogToFile(message1); err != nil {
		t.Fatalf("Expected SaveLogToFile to succeed, got: %v", err)
	}

	dateStr := time.Now().Format("2006-01-02")
	expectedFilePath1 := filepath.Join(tempDir, fmt.Sprintf("ostenia-%s-01.jsonl", dateStr))
	verifyLogFileContentAndLines(t, expectedFilePath1, message1, 1)

	// Verify JSON structure
	data, err := os.ReadFile(expectedFilePath1)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	var entry LogEntry
	if err := json.Unmarshal(data[:len(data)-1], &entry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}
	if entry.Msg != message1 {
		t.Errorf("Expected Msg %s, got %s", message1, entry.Msg)
	}

	// 2. Write 999 more lines to reach 1000 lines
	writeLogLinesBatch(t, app, 999)
	verifyLogFileContentAndLines(t, expectedFilePath1, "", 1000)

	// 3. Write one more log to trigger rotation to sequence 2
	message2 := "This is a rotated log line"
	if err := app.SaveLogToFile(message2); err != nil {
		t.Fatalf("Expected rotating SaveLogToFile to succeed, got: %v", err)
	}

	expectedFilePath2 := filepath.Join(tempDir, fmt.Sprintf("ostenia-%s-02.jsonl", dateStr))
	verifyLogFileContentAndLines(t, expectedFilePath2, message2, 1)
	verifyLogFileContentAndLines(t, expectedFilePath1, "", 1000)
}

func TestApp_SaveLogEntry_Structured(t *testing.T) {
	tempDir := t.TempDir()
	oldLogsDir := os.Getenv("OSTENIA_LOGS_DIR")
	os.Setenv("OSTENIA_LOGS_DIR", tempDir)
	defer os.Setenv("OSTENIA_LOGS_DIR", oldLogsDir)

	app := NewApp()

	entry := LogEntry{
		ID:           "test-id-123",
		Msg:          "[ERR] database connection failed",
		Type:         "error",
		IsServiceLog: false,
		Caller: &CallerInfo{
			FunctionName: "connectDB",
			FileName:     "db.ts",
			Line:         "42",
			Column:       "10",
		},
		RawStack: "Error: database connection failed\n    at connectDB (db.ts:42:10)",
	}

	if err := app.SaveLogEntry(entry); err != nil {
		t.Fatalf("SaveLogEntry failed: %v", err)
	}

	logs, err := app.GetRecentLogs(10)
	if err != nil {
		t.Fatalf("GetRecentLogs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}
	loaded := logs[0]
	if loaded.ID != "test-id-123" {
		t.Errorf("Expected ID test-id-123, got %s", loaded.ID)
	}
	if loaded.Type != "error" {
		t.Errorf("Expected Type error, got %s", loaded.Type)
	}
	if loaded.Caller == nil || loaded.Caller.FunctionName != "connectDB" {
		t.Errorf("Expected Caller function connectDB, got %+v", loaded.Caller)
	}
	if loaded.RawStack != entry.RawStack {
		t.Errorf("Expected RawStack match, got %s", loaded.RawStack)
	}
}

func TestApp_GetRecentLogs_OrderingAndLimit(t *testing.T) {
	tempDir := t.TempDir()
	oldLogsDir := os.Getenv("OSTENIA_LOGS_DIR")
	os.Setenv("OSTENIA_LOGS_DIR", tempDir)
	defer os.Setenv("OSTENIA_LOGS_DIR", oldLogsDir)

	app := NewApp()

	// Initially empty
	emptyLogs, err := app.GetRecentLogs(10)
	if err != nil {
		t.Fatalf("GetRecentLogs on empty dir failed: %v", err)
	}
	if len(emptyLogs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(emptyLogs))
	}

	// Add 5 logs
	for i := 1; i <= 5; i++ {
		_ = app.SaveLogEntry(LogEntry{
			ID:   fmt.Sprintf("id-%d", i),
			Msg:  fmt.Sprintf("Message %d", i),
			Type: "info",
		})
	}

	// Get with limit 3 (should return newest first: 5, 4, 3)
	recent3, err := app.GetRecentLogs(3)
	if err != nil {
		t.Fatalf("GetRecentLogs failed: %v", err)
	}
	if len(recent3) != 3 {
		t.Fatalf("Expected 3 logs, got %d", len(recent3))
	}
	if recent3[0].ID != "id-5" || recent3[1].ID != "id-4" || recent3[2].ID != "id-3" {
		t.Errorf("Unexpected ordering in recent logs: %v, %v, %v", recent3[0].ID, recent3[1].ID, recent3[2].ID)
	}

	// Get with negative/zero limit defaults to 200
	allLogs, err := app.GetRecentLogs(0)
	if err != nil {
		t.Fatalf("GetRecentLogs with 0 limit failed: %v", err)
	}
	if len(allLogs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(allLogs))
	}
}

func TestApp_ClearLogs(t *testing.T) {
	tempDir := t.TempDir()
	oldLogsDir := os.Getenv("OSTENIA_LOGS_DIR")
	os.Setenv("OSTENIA_LOGS_DIR", tempDir)
	defer os.Setenv("OSTENIA_LOGS_DIR", oldLogsDir)

	app := NewApp()

	_ = app.SaveLogToFile("Log message 1")
	_ = app.SaveLogToFile("Log message 2")

	logs, err := app.GetRecentLogs(10)
	if err != nil || len(logs) != 2 {
		t.Fatalf("Expected 2 logs before clear, got %d (err: %v)", len(logs), err)
	}

	if err := app.ClearLogs(); err != nil {
		t.Fatalf("ClearLogs failed: %v", err)
	}

	logsAfter, err := app.GetRecentLogs(10)
	if err != nil || len(logsAfter) != 0 {
		t.Fatalf("Expected 0 logs after clear, got %d (err: %v)", len(logsAfter), err)
	}
}

func TestApp_OpenLogsFolder(t *testing.T) {
	tempDir := t.TempDir()
	oldLogsDir := os.Getenv("OSTENIA_LOGS_DIR")
	os.Setenv("OSTENIA_LOGS_DIR", tempDir)
	defer os.Setenv("OSTENIA_LOGS_DIR", oldLogsDir)

	oldGOOS := service.RuntimeGOOS
	service.RuntimeGOOS = "windows"
	defer func() { service.RuntimeGOOS = oldGOOS }()

	app := NewApp()
	_ = app.OpenLogsFolder()

	// Verify directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Expected logs directory to exist at %s", tempDir)
	}
}

func TestApp_PruneOldLogs(t *testing.T) {
	tempDir := t.TempDir()

	oldFile := filepath.Join(tempDir, "ostenia-2020-01-01-01.jsonl")
	newFile := filepath.Join(tempDir, "ostenia-2099-01-01-01.jsonl")

	if err := os.WriteFile(oldFile, []byte("old\n"), 0644); err != nil {
		t.Fatalf("Failed to write oldFile: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new\n"), 0644); err != nil {
		t.Fatalf("Failed to write newFile: %v", err)
	}

	// Change mtime of oldFile to 30 days ago
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	_ = os.Chtimes(oldFile, thirtyDaysAgo, thirtyDaysAgo)

	pruneOldLogs(tempDir, 14)

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("Expected oldFile to be pruned")
	}
	if _, err := os.Stat(newFile); os.IsNotExist(err) {
		t.Errorf("Expected newFile to remain")
	}
}
