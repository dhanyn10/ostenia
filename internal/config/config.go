package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	WWWRoot     string `json:"wwwRoot"`
	PHPVersion  string `json:"phpVersion"`
	NodeVersion string `json:"nodeVersion"`
}

func GetBaseDir() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}

func LoadConfig() (*Config, error) {
	baseDir := GetBaseDir()
	configPath := filepath.Join(baseDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Default config
		cfg := &Config{
			WWWRoot: filepath.Join(baseDir, "www"),
		}
		os.MkdirAll(cfg.WWWRoot, 0755)
		SaveConfig(cfg)
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func SaveConfig(cfg *Config) error {
	baseDir := GetBaseDir()
	configPath := filepath.Join(baseDir, "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
