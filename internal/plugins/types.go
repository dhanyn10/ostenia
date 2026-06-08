// Package plugins provides functionality for managing external components,
// including discovery, downloading, installation, and version management.
package plugins

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
