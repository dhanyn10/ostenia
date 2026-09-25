package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDefaultPluginsFolderAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ostenia_json_plugins_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", origHome)

	// 1. Test creation of default plugin JSON files
	pluginsDir := EnsureDefaultPluginsFolder()
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		t.Errorf("Expected plugins directory to be created, got not exist")
	}

	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		t.Fatalf("Failed to read plugins dir: %v", err)
	}
	if len(files) == 0 {
		t.Errorf("Expected default plugin json files to be created, got 0 files")
	}

	// 2. Test loading default plugins
	tasks := LoadPluginsFromJSON()
	if len(tasks) == 0 {
		t.Errorf("Expected tasks to be loaded, got 0")
	}

	// Check if PHP is present
	foundPHP := false
	for _, task := range tasks {
		if task.Name == "PHP" {
			foundPHP = true
			if len(task.Versions) == 0 {
				t.Errorf("Expected PHP versions to be populated")
			}
			break
		}
	}
	if !foundPHP {
		t.Errorf("Expected PHP task in loaded plugins")
	}

	// 3. Test custom plugin JSON file
	customPlugin := PluginJSON{
		Name:         "CustomTool",
		Category:     "customtool",
		TargetPrefix: "customtool/customtool-",
		CheckFile:    "custom.exe",
		Versions:     []string{"1.0.0"},
		VersionUrls: map[string]string{
			"1.0.0": "https://example.com/customtool-1.0.0.zip",
		},
	}

	customData, err := json.MarshalIndent(customPlugin, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal custom plugin: %v", err)
	}

	err = os.WriteFile(filepath.Join(pluginsDir, "customtool.json"), customData, 0644)
	if err != nil {
		t.Fatalf("Failed to write custom plugin file: %v", err)
	}

	tasksWithCustom := LoadPluginsFromJSON()
	foundCustom := false
	for _, task := range tasksWithCustom {
		if task.Name == "CustomTool" {
			foundCustom = true
			if task.Version != "1.0.0" {
				t.Errorf("Expected custom plugin version 1.0.0, got %s", task.Version)
			}
			if task.URL != "https://example.com/customtool-1.0.0.zip" {
				t.Errorf("Expected custom plugin URL, got %s", task.URL)
			}
		}
	}
	if !foundCustom {
		t.Errorf("Expected CustomTool task in reloaded plugins")
	}
}
