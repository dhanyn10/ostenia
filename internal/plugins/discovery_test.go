package plugins

import (
	"testing"
)

func TestGetLatestKnownVersions(t *testing.T) {
	tasks := GetLatestKnownVersions()
	if len(tasks) == 0 {
		t.Fatal("Expected at least one task")
	}

	for _, task := range tasks {
		t.Run(task.Name, func(t *testing.T) {
			if len(task.Versions) == 0 {
				t.Errorf("No versions detected for %s", task.Name)
			}
			if task.IconSVG == "" {
				t.Errorf("No icon detected for %s", task.Name)
			}
			for _, v := range task.Versions {
				if task.VersionUrls[v] == "" {
					t.Errorf("No URL for version %s of %s", v, task.Name)
				}
			}

			// Specific checks
			if task.Name == "PHP" {
				foundComposer := false
				for _, mod := range task.Modules {
					if mod.Name == "Composer" {
						foundComposer = true
					}
				}
				if !foundComposer {
					t.Error("Expected Composer module for PHP")
				}
			}
			if task.Name == "Python" {
				if len(task.Modules) != 0 {
					t.Error("Expected no modules for Python")
				}
			}
		})
	}
}

func TestHandleHeidiSQLDetection(t *testing.T) {
	task := &DownloadTask{Name: "HeidiSQL"}
	handleHeidiSQLDetection(task)
}

func TestCheckFileExists(t *testing.T) {
	if checkFileExists("NonExistent", "/tmp", "none.exe") {
		t.Error("Expected false for non-existent file")
	}
}
