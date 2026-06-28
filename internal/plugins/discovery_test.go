package plugins

import (
	"testing"
)

func TestGetLatestKnownVersions(t *testing.T) {
	// Startup usually sets things up, but we can just call it
	tasks := GetLatestKnownVersions()
	if len(tasks) == 0 {
		t.Error("Expected at least one task")
	}

	foundPHP := false
	for _, task := range tasks {
		if task.Name == "PHP" {
			foundPHP = true
			break
		}
	}
	if !foundPHP {
		t.Error("PHP task not found")
	}
}

func TestHandleHeidiSQLDetection(t *testing.T) {
	task := &DownloadTask{Name: "HeidiSQL"}
	handleHeidiSQLDetection(task)
	// Even if not installed on CI, it should run without error
}

func TestCheckFileExists(t *testing.T) {
	if checkFileExists("NonExistent", "/tmp", "none.exe") {
		t.Error("Expected false for non-existent file")
	}
}
