package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"ostenia/internal/config"
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

	// Set the environment variable to point to our temp base directory
	oldHome := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", oldHome)

	// Set custom config file path to avoid pollution
	configPath := filepath.Join(tempDir, "config.json")
	oldConfig := config.SetConfigFile(configPath)
	defer config.SetConfigFile(oldConfig)

	// Load config to initialize config state
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	cfg.BaseDir = tempDir

	app := NewApp()
	app.cfg = cfg

	// 1. Write first log message
	message1 := "First test log message"
	if err := app.SaveLogToFile(message1); err != nil {
		t.Fatalf("Expected SaveLogToFile to succeed, got: %v", err)
	}

	dateStr := time.Now().Format("020106")
	expectedFilePath1 := filepath.Join(tempDir, fmt.Sprintf("%s-01.log", dateStr))
	verifyLogFileContentAndLines(t, expectedFilePath1, message1, 1)

	// 2. Write 999 more lines to reach 1000 lines
	writeLogLinesBatch(t, app, 999)
	verifyLogFileContentAndLines(t, expectedFilePath1, "", 1000)

	// 3. Write one more log to trigger rotation to sequence 2
	message2 := "This is a rotated log line"
	if err := app.SaveLogToFile(message2); err != nil {
		t.Fatalf("Expected rotating SaveLogToFile to succeed, got: %v", err)
	}

	expectedFilePath2 := filepath.Join(tempDir, fmt.Sprintf("%s-02.log", dateStr))
	verifyLogFileContentAndLines(t, expectedFilePath2, message2, 1)
	verifyLogFileContentAndLines(t, expectedFilePath1, "", 1000)
}
