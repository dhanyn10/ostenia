package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	AppVersion    string         `json:"appVersion,omitempty"`
}

var globalConfig *Config
var configPathOverride string
var embeddedWailsJSON []byte

// SetEmbeddedWailsJSON stores the embedded wails.json content for version resolution.
func SetEmbeddedWailsJSON(data []byte) {
	embeddedWailsJSON = data
}

type wailsConfig struct {
	Version string `json:"version"`
	Info    struct {
		ProductVersion string `json:"productVersion"`
	} `json:"info"`
}

// GetAppVersion resolves the application version dynamically from config.json or wails.json.
func GetAppVersion() string {
	// 1. Check if version is explicitly configured in memory
	if globalConfig != nil && globalConfig.AppVersion != "" {
		return FormatVersion(globalConfig.AppVersion)
	}

	// 2. Try reading wails.json from disk (checking baseDir, cwd, parent dirs)
	dirsToTry := []string{
		GetBaseDir(),
		".",
		"..",
		"../..",
	}
	if exe, err := os.Executable(); err == nil {
		dirsToTry = append(dirsToTry, filepath.Dir(exe), filepath.Dir(filepath.Dir(exe)))
	}

	for _, dir := range dirsToTry {
		wailsPath := filepath.Join(dir, "wails.json")
		if data, err := os.ReadFile(wailsPath); err == nil {
			if v := extractVersionFromWailsJSON(data); v != "" {
				return FormatVersion(v)
			}
		}
	}

	// 3. Fallback to embedded wails.json if available
	if len(embeddedWailsJSON) > 0 {
		if v := extractVersionFromWailsJSON(embeddedWailsJSON); v != "" {
			return FormatVersion(v)
		}
	}

	return "v0.10.0"
}

func extractVersionFromWailsJSON(data []byte) string {
	var wc wailsConfig
	if err := json.Unmarshal(data, &wc); err == nil {
		if wc.Info.ProductVersion != "" {
			return wc.Info.ProductVersion
		}
		if wc.Version != "" {
			return wc.Version
		}
	}
	return ""
}

// FormatVersion ensures the version string has a leading 'v'.
func FormatVersion(v string) string {
	v = strings.TrimSpace(v)
	if !strings.HasPrefix(v, "v") && !strings.HasPrefix(v, "V") && v != "" {
		return "v" + v
	}
	return v
}

// SetConfigFile allows overriding the default config.json path for testing purposes.
// It returns the previous config file path override.
func SetConfigFile(path string) string {
	old := configPathOverride
	configPathOverride = path
	return old
}

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
	if envDir := os.Getenv("OSTENIA_HOME"); envDir != "" {
		return filepath.Join(envDir, "config.json")
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
