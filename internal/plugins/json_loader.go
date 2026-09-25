package plugins

import (
	"encoding/json"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/plugins/apache"
	"ostenia/internal/plugins/heidisql"
	"ostenia/internal/plugins/mysql"
	"ostenia/internal/plugins/nginx"
	"ostenia/internal/plugins/nodejs"
	"ostenia/internal/plugins/openssl"
	"ostenia/internal/plugins/php"
	"ostenia/internal/plugins/python"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"sort"
	"strings"
)

// ModuleJSON represents a sub-module configuration in JSON.
type ModuleJSON struct {
	Name      string `json:"name"`
	CheckFile string `json:"checkFile"`
}

// PluginJSON represents the structure of a plugin manifest file.
type PluginJSON struct {
	Name         string            `json:"name"`
	Category     string            `json:"category"`
	TargetPrefix string            `json:"targetPrefix"`
	CheckFile    string            `json:"checkFile"`
	IconSVG      string            `json:"iconSVG,omitempty"`
	Info         string            `json:"info,omitempty"`
	Versions     []string          `json:"versions"`
	VersionUrls  map[string]string `json:"versionUrls"`
	Modules      []ModuleJSON      `json:"modules,omitempty"`
}

// GetPluginsDir returns the path to the plugins JSON directory.
func GetPluginsDir() string {
	return filepath.Join(config.GetBaseDir(), "plugins")
}

// getDefaultPluginJSONs returns the default manifests for built-in plugins.
func getDefaultPluginJSONs() map[string]PluginJSON {
	return map[string]PluginJSON{
		"php.json": {
			Name:         "PHP",
			Category:     "php",
			TargetPrefix: "php/php-",
			CheckFile:    "php.exe",
			IconSVG:      php.GetIcon(),
			Info:         "PHP Hypertext Preprocessor",
			Versions:     []string{"8.3.3", "8.2.16"},
			VersionUrls: map[string]string{
				"8.3.3":  "https://windows.php.net/downloads/releases/php-8.3.3-Win32-vs16-x64.zip",
				"8.2.16": "https://windows.php.net/downloads/releases/php-8.2.16-Win32-vs16-x64.zip",
			},
			Modules: []ModuleJSON{
				{Name: "Composer", CheckFile: "composer.phar"},
			},
		},
		"apache.json": {
			Name:         "Apache",
			Category:     "apache",
			TargetPrefix: "apache/httpd-",
			CheckFile:    "bin/httpd.exe",
			IconSVG:      apache.GetIcon(),
			Versions:     []string{"2.4.58"},
			VersionUrls: map[string]string{
				"2.4.58": "https://www.apachelounge.com/download/VS17/binaries/httpd-2.4.58-240131-win64-VS17.zip",
			},
		},
		"mysql.json": {
			Name:         "MySQL",
			Category:     "mysql",
			TargetPrefix: "mysql/mysql-",
			CheckFile:    "bin/mysqld.exe",
			IconSVG:      mysql.GetIcon(),
			Versions:     []string{"8.0.36"},
			VersionUrls: map[string]string{
				"8.0.36": "https://cdn.mysql.com/Downloads/MySQL-8.0/mysql-8.0.36-winx64.zip",
			},
		},
		"nodejs.json": {
			Name:         "Node.js",
			Category:     "nodejs",
			TargetPrefix: "nodejs/node-v",
			CheckFile:    "node.exe",
			IconSVG:      nodejs.GetIcon(),
			Versions:     []string{"20.11.1"},
			VersionUrls: map[string]string{
				"20.11.1": "https://nodejs.org/dist/v20.11.1/node-v20.11.1-win-x64.zip",
			},
		},
		"python.json": {
			Name:         "Python",
			Category:     "python",
			TargetPrefix: "python/python-",
			CheckFile:    "python.exe",
			IconSVG:      python.GetIcon(),
			Versions:     []string{"3.12.2"},
			VersionUrls: map[string]string{
				"3.12.2": "https://www.python.org/ftp/python/3.12.2/python-3.12.2-embed-amd64.zip",
			},
		},
		"heidisql.json": {
			Name:         "HeidiSQL",
			Category:     "heidisql",
			TargetPrefix: "heidisql/heidisql-",
			CheckFile:    "heidisql.exe",
			IconSVG:      heidisql.GetIcon(),
			Versions:     []string{"12.6.0.6765"},
			VersionUrls: map[string]string{
				"12.6.0.6765": "https://www.heidisql.com/downloads/releases/HeidiSQL_12.6.0.6765_64_Portable.zip",
			},
		},
		"nginx.json": {
			Name:         "Nginx",
			Category:     "nginx",
			TargetPrefix: "nginx/nginx-",
			CheckFile:    "nginx.exe",
			IconSVG:      nginx.GetIcon(),
			Versions:     []string{"1.24.0"},
			VersionUrls: map[string]string{
				"1.24.0": "https://nginx.org/download/nginx-1.24.0.zip",
			},
		},
		"openssl.json": {
			Name:         "OpenSSL",
			Category:     "openssl",
			TargetPrefix: "openssl/openssl-",
			CheckFile:    "bin/openssl.exe",
			IconSVG:      openssl.GetIcon(),
			Versions:     []string{"3.2.1"},
			VersionUrls: map[string]string{
				"3.2.1": "https://slproweb.com/download/Win64OpenSSL-3_2_1.exe",
			},
		},
	}
}

// EnsureDefaultPluginsFolder initializes the plugins directory with default manifests if missing.
func EnsureDefaultPluginsFolder() string {
	pluginsDir := GetPluginsDir()
	_ = os.MkdirAll(pluginsDir, 0755)

	defaults := getDefaultPluginJSONs()
	for filename, pluginJSON := range defaults {
		filePath := filepath.Join(pluginsDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			data, err := json.MarshalIndent(pluginJSON, "", "  ")
			if err == nil {
				_ = os.WriteFile(filePath, data, 0644)
			}
		}
	}
	return pluginsDir
}

// LoadPluginsFromJSON scans the plugins directory and parses all JSON manifest files.
func LoadPluginsFromJSON() []DownloadTask {
	pluginsDir := EnsureDefaultPluginsFolder()

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil
	}

	baseDir := config.GetBaseDir()
	var tasks []DownloadTask

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		filePath := filepath.Join(pluginsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var pj PluginJSON
		if err := json.Unmarshal(data, &pj); err != nil {
			continue
		}

		if pj.Name == "" {
			pj.Name = strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		}

		task := buildTaskFromJSON(pj, baseDir)
		tasks = append(tasks, task)
	}

	// Sort tasks by name for deterministic order
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Name < tasks[j].Name
	})

	return tasks
}

func getFallbackIcon(category string) string {
	switch strings.ToLower(category) {
	case "php":
		return php.GetIcon()
	case "apache":
		return apache.GetIcon()
	case "mysql":
		return mysql.GetIcon()
	case "nodejs":
		return nodejs.GetIcon()
	case "python":
		return python.GetIcon()
	case "heidisql":
		return heidisql.GetIcon()
	case "nginx":
		return nginx.GetIcon()
	case "openssl":
		return openssl.GetIcon()
	default:
		return ""
	}
}

func buildTaskFromJSON(pj PluginJSON, baseDir string) DownloadTask {
	icon := pj.IconSVG
	if icon == "" {
		icon = getFallbackIcon(pj.Category)
	}

	// Fallback dynamic version detection if JSON has no versions
	vers := pj.Versions
	urls := pj.VersionUrls
	if len(vers) == 0 && urls == nil {
		vers, urls = dynamicDetectFallback(pj.Category)
	}

	t := DownloadTask{
		Name:        pj.Name,
		CheckFile:   pj.CheckFile,
		IconSVG:     icon,
		VersionUrls: urls,
		Versions:    vers,
		Info:        pj.Info,
	}

	if len(vers) > 0 {
		t.Version = vers[0]
		if urls != nil {
			t.URL = urls[vers[0]]
		}
		t.Target = pj.TargetPrefix + vers[0]
	}

	// 1. Detect ALL installed versions
	installedMap := utils.GetInstalledVersionPaths(baseDir, pj.Category, t.CheckFile)
	t.InstalledVers = make([]string, 0, len(installedMap))
	for v := range installedMap {
		t.InstalledVers = append(t.InstalledVers, v)
	}
	sort.Strings(t.InstalledVers)

	// Special Cases
	if strings.EqualFold(pj.Name, "HeidiSQL") {
		handleHeidiSQLDetection(&t)
		return t
	}
	if strings.EqualFold(pj.Name, "OpenSSL") {
		handleOpenSSLDetection(&t)
		return t
	}

	currentPath := filepath.Join(baseDir, "bin", pj.Category, "current")

	// Detect modules
	for _, modJSON := range pj.Modules {
		isModInstalled := false
		if _, err := os.Stat(filepath.Join(currentPath, modJSON.CheckFile)); err == nil {
			isModInstalled = true
		}
		status := "Not Installed"
		version := ""
		if isModInstalled {
			status = "Ready"
			version = getModuleVersion(pj.Category, modJSON.Name, currentPath)
		}
		t.Modules = append(t.Modules, PluginModule{
			Name:        modJSON.Name,
			IsInstalled: isModInstalled,
			Status:      status,
			Version:     version,
			CheckFile:   modJSON.CheckFile,
		})
	}

	if pj.Info == "" {
		if strings.EqualFold(pj.Category, "python") {
			t.Info = python.GetInfo(currentPath)
		}
	}

	// Check if current junction/symlink is active
	if checkCurrentFunctionality(&t, currentPath, baseDir) {
		t.IsInstalled = true
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
			activeFolder := filepath.Base(resolved)
			if activeFolder != "current" {
				activeVer := utils.NormalizeVersion(activeFolder)
				if activeVer != "current" {
					t.ActiveVersion = activeVer
				}
			}
		}
	}

	return t
}

func dynamicDetectFallback(category string) ([]string, map[string]string) {
	switch strings.ToLower(category) {
	case "php":
		return php.DetectVersions()
	case "apache":
		return apache.DetectVersions()
	case "mysql":
		return mysql.DetectVersions()
	case "nodejs":
		return nodejs.DetectVersions()
	case "python":
		return python.DetectVersions()
	case "heidisql":
		return heidisql.DetectVersions()
	case "nginx":
		return nginx.DetectVersions()
	case "openssl":
		return openssl.DetectVersions()
	default:
		return nil, nil
	}
}

func getModuleVersion(category, moduleName, currentPath string) string {
	if strings.EqualFold(category, "php") {
		return php.GetModuleVersion(moduleName, currentPath)
	}
	if strings.EqualFold(category, "python") {
		return python.GetModuleVersion(moduleName, currentPath)
	}
	return ""
}
