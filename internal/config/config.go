package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the global application configuration settings
type Config struct {
	BaseDir       string         `json:"baseDir"` // Root directory for the whole Ostenia environment
	WWWRoot       string         `json:"wwwRoot"`
	PHPVersion    string         `json:"phpVersion"`
	NodeVersion   string         `json:"nodeVersion"`
	ApacheHTTPS   bool           `json:"apacheHttps"`
	NginxHTTPS    bool           `json:"nginxHttps"`
	Proxies       map[string]int `json:"proxies"` // folder_name -> target_port
	DefaultEditor string         `json:"defaultEditor"`
}

var globalConfig *Config
var configPathOverride string

// GetBaseDir returns the root directory where Ostenia apps and binaries are stored.
// It prioritizes the explicitly set BaseDir in config, then OSTENIA_HOME environment variable,
// and finally falls back to the directory of the application executable.
func GetBaseDir() string {
	// If a custom base dir is set in config, use it
	if globalConfig != nil && globalConfig.BaseDir != "" {
		return globalConfig.BaseDir
	}

	if envDir := os.Getenv("OSTENIA_HOME"); envDir != "" {
		return envDir
	}
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}

func getConfigPath() string {
	if configPathOverride != "" {
		return configPathOverride
	}
	exePath, _ := os.Executable()
	return filepath.Join(filepath.Dir(exePath), "config.json")
}

// LoadConfig reads the application configuration from config.json.
// If the file doesn't exist, it creates a default configuration.
func LoadConfig() (*Config, error) {
	// 1. Find the local config.json (relative to executable)
	localConfigPath := getConfigPath()

	if _, err := os.Stat(localConfigPath); os.IsNotExist(err) {
		// Default config if none exists
		baseDir := filepath.Dir(localConfigPath)
		cfg := &Config{
			BaseDir:     baseDir,
			WWWRoot:     filepath.Join(baseDir, "www"),
			ApacheHTTPS: false,
			NginxHTTPS:  false,
			Proxies:     make(map[string]int),
		}
		_ = os.MkdirAll(cfg.WWWRoot, 0755)
		_ = SaveConfig(cfg)
		globalConfig = cfg
		return cfg, nil
	}

	data, err := os.ReadFile(localConfigPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err == nil {
		if cfg.Proxies == nil {
			cfg.Proxies = make(map[string]int)
		}
		globalConfig = &cfg
	}
	return &cfg, err
}

// SaveConfig persists the provided configuration to config.json
func SaveConfig(cfg *Config) error {
	configPath := getConfigPath()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	globalConfig = cfg
	return os.WriteFile(configPath, data, 0644)
}
