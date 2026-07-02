package interfaces

import (
	"ostenia/internal/config"
)

type ServiceDetailedInfo struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	Port          int    `json:"port"`
	Ports         []int  `json:"ports"`
	RemainingDays int    `json:"remainingDays,omitempty"`
	ActiveVersion string `json:"activeVersion,omitempty"`
}

type Orchestrator interface {
	SetRuntime(r Runtime)
	SetActiveTab(tab string)
	RequestRefresh()
	StartWatcher()
	IsRunning(name string) bool
	GetDetailedInfo(name string) ServiceDetailedInfo
	StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error
	StartService(name string, binaryPath string, args []string, workingDir string) error
	StopService(name string) error
	StopAll()
}

// DownloadTask represents a plugin's metadata and state.
type DownloadTask struct {
	Name          string            `json:"name"`          // e.g., "PHP", "Node.js"
	URL           string            `json:"url"`           // Default download URL
	Version       string            `json:"version"`       // Selected or latest version
	Versions      []string          `json:"versions"`      // All available versions from remote
	VersionUrls   map[string]string `json:"versionUrls"`   // Map of version -> download URL
	InstalledVers []string          `json:"installedVers"` // Versions found on disk
	Target        string            `json:"target"`        // Relative installation path (e.g., "php/php-8.2.0")
	CheckFile     string            `json:"checkFile"`     // Executable to verify (e.g., "php.exe")
	IsInstalled   bool              `json:"isInstalled"`   // True if the 'current' symlink is valid
	IconSVG       string            `json:"iconSvg"`       // Icon for UI
	Info          string            `json:"info"`          // Additional info (e.g., "Pip 24.0")
	Modules       []PluginModule    `json:"modules"`       // Sub-plugins/modules (e.g., Composer, Pip)
}

type PluginModule struct {
	Name        string `json:"name"`
	IsInstalled bool   `json:"isInstalled"`
	Status      string `json:"status"` // e.g., "Not Installed", "Ready", "Installing..."
	Version     string `json:"version"`
	CheckFile   string `json:"-"`
}

// Progress reports download/extraction status to frontend.
type Progress struct {
	Name       string  `json:"name"`       // Name of the task
	Percentage float64 `json:"percentage"` // Current progress percentage (0-100)
	Status     string  `json:"status"`     // Text description (e.g., "Downloading...", "Extracting...")
	Speed      string  `json:"speed"`      // Human-readable speed (e.g., "2.5 MB/s")
	Downloaded string  `json:"downloaded"` // Human-readable size downloaded (e.g., "10.2 MB")
}

type PluginManager interface {
	DownloadAndExtract(task DownloadTask) error
	DeleteVersion(category, version string) error
	CancelDownload(category string)
	GetInstalledVersionPaths(category, checkFile string) map[string]string
	InstallModule(moduleName string, phpPath string, emitProgress func(string, float64, string)) error
	UninstallModule(moduleName string, phpPath string) error
}

type RemoteFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime int64  `json:"modTime"`
	Mode    string `json:"mode"`
}

type SSLManager interface {
	GenerateRootCA(destDir string) error
	GetRemainingDays(certPath string) (int, error)
	SignCertificate(caDir string, domain string, destDir string) error
}

type SSHManager interface {
	GetSessions() ([]config.SSHSession, error)
	SaveSessions(sessions []config.SSHSession) error
	Connect(session config.SSHSession) error
	Disconnect(sessionID string)
	SendInput(sessionID string, input string) error
	ResizeTerminal(sessionID string, cols, rows int) error
	ListFiles(sessionID string, path string) ([]RemoteFile, error)
	ExecuteSFTPAction(sessionID string, action, path, newPath string) error
	DownloadFile(sessionID string, remotePath string, localPath string) error
	UploadFile(sessionID string, localPath string, remotePath string) error
	EditFile(sessionID string, remotePath string, editor string) error
	GetCurrentPath(sessionID string) (string, error)
}
