package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	BaseDir       string         `json:"baseDir"` // New: Root directory for the whole Ostenia environment
	WWWRoot       string         `json:"wwwRoot"`
	PHPVersion    string         `json:"phpVersion"`
	NodeVersion   string         `json:"nodeVersion"`
	ApacheHTTPS   bool           `json:"apacheHttps"`
	NginxHTTPS    bool           `json:"nginxHttps"`
	Proxies       map[string]int `json:"proxies"` // folder_name -> target_port
	DefaultEditor string         `json:"defaultEditor"`
}

var globalConfig *Config

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

func LoadConfig() (*Config, error) {
	// 1. Find the local config.json (relative to executable)
	exePath, _ := os.Executable()
	localConfigPath := filepath.Join(filepath.Dir(exePath), "config.json")

	if _, err := os.Stat(localConfigPath); os.IsNotExist(err) {
		// Default config if none exists
		cfg := &Config{
			BaseDir:     filepath.Dir(exePath),
			WWWRoot:     filepath.Join(filepath.Dir(exePath), "www"),
			ApacheHTTPS: false,
			NginxHTTPS:  false,
			Proxies:     make(map[string]int),
		}
		os.MkdirAll(cfg.WWWRoot, 0755)
		SaveConfig(cfg)
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

func SaveConfig(cfg *Config) error {
	exePath, _ := os.Executable()
	configPath := filepath.Join(filepath.Dir(exePath), "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	globalConfig = cfg
	return os.WriteFile(configPath, data, 0644)
}
