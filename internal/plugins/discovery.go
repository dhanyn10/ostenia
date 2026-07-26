package plugins

import (
	"os"
	"strings"
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
)

type pluginDefinition struct {
	Name             string
	Category         string
	TargetPrefix     string
	CheckFile        string
	Detect           func() ([]string, map[string]string)
	GetIcon          func() string
	GetInfo          func(path string) string
	GetModules       func() []utils.ModuleDefinition
	GetModuleVersion func(name string, path string) string
}

var (
	DetectHeidiSQLInstallationOverride func() (string, string)
	unzipFunc                          = Unzip
)

func DetectHeidiSQLInstallation() (string, string) {
	if DetectHeidiSQLInstallationOverride != nil {
		return DetectHeidiSQLInstallationOverride()
	}
	return utils.DetectHeidiSQLInstallation()
}

func GetLatestKnownVersions() []DownloadTask {
	definitions := []pluginDefinition{
		{
			Name: "PHP", Category: "php", TargetPrefix: "php/php-", CheckFile: "php.exe",
			Detect: php.DetectVersions, GetIcon: php.GetIcon,
			GetModules: php.GetModules, GetModuleVersion: php.GetModuleVersion,
		},
		{
			Name: "Apache", Category: "apache", TargetPrefix: "apache/httpd-", CheckFile: "bin/httpd.exe",
			Detect: apache.DetectVersions, GetIcon: apache.GetIcon,
		},
		{
			Name: "MySQL", Category: "mysql", TargetPrefix: "mysql/mysql-", CheckFile: "bin/mysqld.exe",
			Detect: mysql.DetectVersions, GetIcon: mysql.GetIcon,
		},
		{
			Name: "Node.js", Category: "nodejs", TargetPrefix: "nodejs/node-v", CheckFile: "node.exe",
			Detect: nodejs.DetectVersions, GetIcon: nodejs.GetIcon,
		},
		{
			Name: "Python", Category: "python", TargetPrefix: "python/python-", CheckFile: "python.exe",
			Detect: python.DetectVersions, GetIcon: python.GetIcon,
			GetInfo:    python.GetInfo,
			GetModules: python.GetModules, GetModuleVersion: python.GetModuleVersion,
		},
		{
			Name: "HeidiSQL", Category: "heidisql", TargetPrefix: "heidisql/heidisql-", CheckFile: "heidisql.exe",
			Detect: heidisql.DetectVersions, GetIcon: heidisql.GetIcon,
		},
		{
			Name: "Nginx", Category: "nginx", TargetPrefix: "nginx/nginx-", CheckFile: "nginx.exe",
			Detect: nginx.DetectVersions, GetIcon: nginx.GetIcon,
		},
		{
			Name: "OpenSSL", Category: "openssl", TargetPrefix: "openssl/openssl-", CheckFile: "bin/openssl.exe",
			Detect: openssl.DetectVersions, GetIcon: openssl.GetIcon,
		},
	}

	var tasks []DownloadTask
	baseDir := config.GetBaseDir()

	for _, def := range definitions {
		tasks = append(tasks, createDownloadTask(def, baseDir))
	}

	return tasks
}

func createDownloadTask(def pluginDefinition, baseDir string) DownloadTask {
	vers, urls := def.Detect()
	t := DownloadTask{
		Name:        def.Name,
		CheckFile:   def.CheckFile,
		IconSVG:     def.GetIcon(),
		VersionUrls: urls,
		Versions:    vers,
	}

	if len(vers) > 0 {
		t.Version = vers[0]
		t.URL = urls[vers[0]]
		t.Target = def.TargetPrefix + vers[0]
	}

	// 1. Detect ALL installed versions
	installedMap := utils.GetInstalledVersionPaths(baseDir, def.Category, t.CheckFile)
	t.InstalledVers = make([]string, 0, len(installedMap))
	for v := range installedMap {
		t.InstalledVers = append(t.InstalledVers, v)
	}
	sort.Strings(t.InstalledVers)

	// Handle Special Cases
	if t.Name == "HeidiSQL" {
		handleHeidiSQLDetection(&t)
		return t
	}
	if t.Name == "OpenSSL" {
		handleOpenSSLDetection(&t)
		return t
	}

	// 1.5 Detect Modules and Info
	currentPath := filepath.Join(baseDir, "bin", def.Category, "current")
	detectModules(&t, def, currentPath)

	if def.GetInfo != nil {
		t.Info = def.GetInfo(currentPath)
	}

	// 2. Check if the currently active 'current' link is functional
	if checkCurrentFunctionality(&t, currentPath, baseDir) {
		t.IsInstalled = true
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
			activeFolder := filepath.Base(resolved)
			if activeFolder != "current" {
				activeVer := activeFolder
				activeVer = strings.TrimPrefix(activeVer, "php-")
				activeVer = strings.TrimPrefix(activeVer, "httpd-")
				activeVer = strings.TrimPrefix(activeVer, "mysql-")
				activeVer = strings.TrimPrefix(activeVer, "nginx-")
				activeVer = strings.TrimPrefix(activeVer, "node-v")
				activeVer = strings.TrimPrefix(activeVer, "python-")
				activeVer = strings.TrimPrefix(activeVer, "heidisql-")
				if activeVer != "current" {
					t.ActiveVersion = activeVer
				}
			}
		}
	}

	return t
}

func handleHeidiSQLDetection(t *DownloadTask) {
	exePath, _ := DetectHeidiSQLInstallation()
	if exePath != "" {
		t.IsInstalled = true
		if len(t.InstalledVers) == 0 {
			t.InstalledVers = []string{"System"}
		}
	}
}

func handleOpenSSLDetection(t *DownloadTask) {
	t.InstalledVers = nil
	t.IsInstalled = false
	if gv := openssl.DetectInstalledVersion(); gv != "" {
		t.Version = gv
		t.InstalledVers = []string{gv}
		t.IsInstalled = true
	}
}

func detectModules(t *DownloadTask, def pluginDefinition, currentPath string) {
	if def.GetModules == nil {
		return
	}
	for _, modDef := range def.GetModules() {
		isModInstalled := false
		if _, err := os.Stat(filepath.Join(currentPath, modDef.CheckFile)); err == nil {
			isModInstalled = true
		}
		status := "Not Installed"
		version := ""
		if isModInstalled {
			status = "Ready"
			if def.GetModuleVersion != nil {
				version = def.GetModuleVersion(modDef.Name, currentPath)
			}
		}
		t.Modules = append(t.Modules, PluginModule{
			Name: modDef.Name, IsInstalled: isModInstalled, Status: status, Version: version, CheckFile: modDef.CheckFile,
		})
	}
}

func checkCurrentFunctionality(t *DownloadTask, currentPath string, baseDir string) bool {
	if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		return checkFileExists(t.Name, resolved, t.CheckFile)
	}

	if t.Target != "" {
		targetPath := filepath.Join(baseDir, "bin", t.Target)
		return checkFileExists(t.Name, targetPath, t.CheckFile)
	}

	return false
}

func checkFileExists(pluginName, basePath, checkFile string) bool {
	cf := filepath.Join(basePath, checkFile)
	if pluginName == "Apache" {
		if _, err := os.Stat(cf); os.IsNotExist(err) {
			cf = filepath.Join(basePath, "Apache24", "bin", "httpd.exe")
		}
	}

	_, err := os.Stat(cf)
	return err == nil
}
