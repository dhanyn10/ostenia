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
}

// Progress reports download/extraction status to frontend.
type Progress struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
	Speed      string  `json:"speed"`
	Downloaded string  `json:"downloaded"`
}
