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
	err = app.SaveLogToFile(message1)
	if err != nil {
		t.Fatalf("Expected SaveLogToFile to succeed, got: %v", err)
	}

	dateStr := time.Now().Format("020106")
	expectedFileName1 := fmt.Sprintf("%s-01.log", dateStr)
	expectedFilePath1 := filepath.Join(tempDir, expectedFileName1)

	// Verify that the file exists and contains the message
	content1, err := os.ReadFile(expectedFilePath1)
	if err != nil {
		t.Fatalf("Expected file to exist at %s, got error: %v", expectedFilePath1, err)
	}
	if !strings.Contains(string(content1), message1) {
		t.Errorf("Expected content to contain '%s', got: %s", message1, string(content1))
	}

	// 2. Count current lines (it should be 1)
	lines, err := countLines(expectedFilePath1)
	if err != nil {
		t.Fatalf("Failed to count lines: %v", err)
	}
	if lines != 1 {
		t.Errorf("Expected line count to be 1, got %d", lines)
	}

	// 3. Write 999 more lines to reach 1000 lines
	for i := 0; i < 999; i++ {
		err = app.SaveLogToFile(fmt.Sprintf("Log line %d", i))
		if err != nil {
			t.Fatalf("Failed to write log line at index %d: %v", i, err)
		}
	}

	// Check that the file is still the first sequence file and has exactly 1000 lines
	lines, err = countLines(expectedFilePath1)
	if err != nil {
		t.Fatalf("Failed to count lines: %v", err)
	}
	if lines != 1000 {
		t.Errorf("Expected line count to be 1000, got %d", lines)
	}

	// 4. Write one more log to trigger rotation to sequence 2
	message2 := "This is a rotated log line"
	err = app.SaveLogToFile(message2)
	if err != nil {
		t.Fatalf("Expected rotating SaveLogToFile to succeed, got: %v", err)
	}

	expectedFileName2 := fmt.Sprintf("%s-02.log", dateStr)
	expectedFilePath2 := filepath.Join(tempDir, expectedFileName2)

	// Verify that the rotated file exists and contains message2
	content2, err := os.ReadFile(expectedFilePath2)
	if err != nil {
		t.Fatalf("Expected rotated file to exist at %s, got error: %v", expectedFilePath2, err)
	}
	if !strings.Contains(string(content2), message2) {
		t.Errorf("Expected rotated file content to contain '%s', got: %s", message2, string(content2))
	}

	// Verify that the first file still has exactly 1000 lines
	lines1, _ := countLines(expectedFilePath1)
	if lines1 != 1000 {
		t.Errorf("Expected first file to stay at 1000 lines, got %d", lines1)
	}

	// Verify that the second file has exactly 1 line
	lines2, _ := countLines(expectedFilePath2)
	if lines2 != 1 {
		t.Errorf("Expected second file to have exactly 1 line, got %d", lines2)
	}
}
