package domain

import (
	"context"
	"time"
)

// Config represents the global application configuration settings.
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

// SSHSession defines the metadata and credentials required to connect to a remote host.
type SSHSession struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	AuthMethod string `json:"authMethod"` // "password" or "key"
	Password   string `json:"password,omitempty"`
	KeyPath    string `json:"keyPath,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
	LastPath   string `json:"lastPath,omitempty"`
	CreatedAt  int64  `json:"createdAt"`
	MaxTimeout int    `json:"maxTimeout,omitempty"`
	MaxRetries int    `json:"maxRetries,omitempty"`
}

// RemoteFileInfo represents file or directory metadata from a remote host.
type RemoteFileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	Mode    string    `json:"mode"`
}

// NetworkHost represents an entry in the system hosts file.
type NetworkHost struct {
	IP    string   `json:"ip"`
	Names []string `json:"names"`
}

// DownloadTask represents an available plugin or version to download.
type DownloadTask struct {
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	URL           string   `json:"url"`
	FileName      string   `json:"fileName"`
	ExtractedDir  string   `json:"extractedDir"`
	ActiveVersion string   `json:"activeVersion,omitempty"`
	Binaries      []string `json:"binaries,omitempty"`
}

// ServiceDetailedInfo contains status and metadata for a service daemon or binary.
type ServiceDetailedInfo struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	Port          int    `json:"port"`
	Ports         []int  `json:"ports"`
	RemainingDays int    `json:"remainingDays,omitempty"`
	ActiveVersion string `json:"activeVersion,omitempty"`
}

// SSHManager defines domain interactions for SSH sessions and SFTP.
type SSHManager interface {
	GetSessions() ([]SSHSession, error)
	SaveSessions(sessions []SSHSession) error
	Connect(ctx context.Context, session SSHSession) error
	GetCurrentPath(sessionID string) (string, error)
	ExecuteSFTPAction(sessionID, action, path, newPath string) error
}

// Orchestrator defines domain interactions for background service lifecycle.
type Orchestrator interface {
	IsRunning(name string) bool
	GetDetailedInfo(name string) ServiceDetailedInfo
	StartService(ctx context.Context, name, binaryPath string, args []string, workingDir string) error
	StopService(ctx context.Context, name string) error
}

// ConfigRepository defines domain operations for loading and persisting configuration.
type ConfigRepository interface {
	LoadConfig() (*Config, error)
	SaveConfig(cfg *Config) error
	GetBaseDir() string
}
