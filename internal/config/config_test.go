package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetBaseDirFromEnv(t *testing.T) {
	tmpDir := t.TempDir()

	// Mock OSTENIA_HOME to control GetBaseDir
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	// Ensure globalConfig is nil to avoid it taking precedence
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() { globalConfig = oldGlobalConfig }()

	if GetBaseDir() != tmpDir {
		t.Errorf("GetBaseDir() = %v, want %v", GetBaseDir(), tmpDir)
	}
}

func TestGetBaseDirFromConfig(t *testing.T) {
	oldGlobalConfig := globalConfig
	globalConfig = &Config{BaseDir: "/custom/dir"}
	defer func() { globalConfig = oldGlobalConfig }()

	if GetBaseDir() != "/custom/dir" {
		t.Errorf("GetBaseDir() = %v, want /custom/dir", GetBaseDir())
	}
}

func TestJSONMarshaling(t *testing.T) {
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
}

func TestLoadAndSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Setup override and reset globalConfig
	oldOverride := configPathOverride
	SetConfigFile(configPath)
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		SetConfigFile(oldOverride)
		globalConfig = oldGlobalConfig
	}()

	// Test default config creation
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed to create default config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config.json was not created")
	}

	if cfg.WWWRoot == "" {
		t.Errorf("Default WWWRoot should not be empty")
	}

	// Test saving modified config
	cfg.PHPVersion = "8.2.0"
	err = SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Test loading again
	cfg2, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed to load existing config: %v", err)
	}

	if cfg2.PHPVersion != "8.2.0" {
		t.Errorf("Loaded config PHPVersion = %v, want 8.2.0", cfg2.PHPVersion)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	err := os.WriteFile(configPath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Setup override and reset globalConfig
	oldOverride := configPathOverride
	SetConfigFile(configPath)
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		SetConfigFile(oldOverride)
		globalConfig = oldGlobalConfig
	}()

	_, err = LoadConfig()
	if err == nil {
		t.Errorf("LoadConfig should have failed with invalid JSON")
	}
}

func TestGetBaseDir_Fallback(t *testing.T) {
	oldHome := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", "")
	defer os.Setenv("OSTENIA_HOME", oldHome)

	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() { globalConfig = oldGlobalConfig }()

	baseDir := GetBaseDir()
	if baseDir == "" {
		t.Error("Expected GetBaseDir() fallback to not be empty")
	}
}

func TestLoadConfig_IOError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	err := os.Mkdir(configPath, 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	oldOverride := configPathOverride
	SetConfigFile(configPath)
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		SetConfigFile(oldOverride)
		globalConfig = oldGlobalConfig
	}()

	_, err = LoadConfig()
	if err == nil {
		t.Error("Expected LoadConfig to fail with a directory instead of file")
	}
}

func TestSaveConfig_MarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "non_existent_subdir", "config.json")

	oldOverride := configPathOverride
	SetConfigFile(configPath)
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		SetConfigFile(oldOverride)
		globalConfig = oldGlobalConfig
	}()

	cfg := &Config{
		BaseDir: tmpDir,
	}
	err := SaveConfig(cfg)
	if err == nil {
		t.Error("Expected SaveConfig to fail when folder does not exist")
	}
}

func TestLoadSSHSessions_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tmpDir)
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		os.Setenv("OSTENIA_HOME", oldHome)
		globalConfig = oldGlobalConfig
	}()

	err := os.WriteFile(filepath.Join(tmpDir, "ssh_sessions.json"), []byte("{invalid"), 0644)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	_, err = LoadSSHSessions()
	if err == nil {
		t.Error("Expected LoadSSHSessions to fail with invalid json")
	}
}

func TestLoadSSHSessions_IOError(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tmpDir)
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		os.Setenv("OSTENIA_HOME", oldHome)
		globalConfig = oldGlobalConfig
	}()

	err := os.Mkdir(filepath.Join(tmpDir, "ssh_sessions.json"), 0755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	_, err = LoadSSHSessions()
	if err == nil {
		t.Error("Expected LoadSSHSessions to fail with folder")
	}
}

func TestSaveSSHSessions_JSONError(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", filepath.Join(tmpDir, "non_existent"))
	oldGlobalConfig := globalConfig
	globalConfig = nil
	defer func() {
		os.Setenv("OSTENIA_HOME", oldHome)
		globalConfig = oldGlobalConfig
	}()

	err := SaveSSHSessions([]SSHSession{{ID: "1"}})
	if err == nil {
		t.Error("Expected SaveSSHSessions to fail with invalid destination path")
	}
}

func TestEncryptDecrypt_AESError(t *testing.T) {
	oldKey := encryptionKey
	encryptionKey = []byte("short")
	defer func() { encryptionKey = oldKey }()

	_, err := Encrypt("hello")
	if err == nil {
		t.Error("Expected Encrypt to fail with short key")
	}

	_, err = Decrypt("hello")
	if err == nil {
		t.Error("Expected Decrypt to fail with short key")
	}
}
