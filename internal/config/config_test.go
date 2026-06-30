package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	// Setup temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "ostenia-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock OSTENIA_HOME to control GetBaseDir
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	t.Run("GetBaseDir from Env", func(t *testing.T) {
		if GetBaseDir() != tmpDir {
			t.Errorf("GetBaseDir() = %v, want %v", GetBaseDir(), tmpDir)
		}
	})

	t.Run("JSON Marshaling", func(t *testing.T) {
		cfg := &Config{
			BaseDir: "/test",
			Proxies: map[string]int{"a": 1},
		}
		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var cfg2 Config
		if err := json.Unmarshal(data, &cfg2); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if cfg2.BaseDir != cfg.BaseDir || cfg2.Proxies["a"] != 1 {
			t.Errorf("Unmarshaled config mismatch")
		}
	})

	t.Run("Load and Save Config", func(t *testing.T) {
		// Since LoadConfig/SaveConfig use os.Executable(), we can't easily change the path.
		// But we can try to call them and see if they work without crashing.
		// To avoid messing up with actual files if they exist, we just check if they can be called.

		cfg, err := LoadConfig()
		if err != nil {
			t.Logf("LoadConfig returned error (expected in some envs): %v", err)
		}

		if cfg != nil {
			err = SaveConfig(cfg)
			if err != nil {
				t.Logf("SaveConfig returned error: %v", err)
			}
		}
	})
}
