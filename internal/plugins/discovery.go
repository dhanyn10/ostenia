package plugins

import (
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

func DetectHeidiSQLInstallation() (string, string) {
	return utils.DetectHeidiSQLInstallation()
}

func getPluginDefinitions() []pluginDefinition {
	return []pluginDefinition{
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
			GetInfo: python.GetInfo,
			GetModules: python.GetModules, GetModuleVersion: python.GetModuleVersion,
		},
		{
			Name: "HeidiSQL", Category: "heidisql", TargetPrefix: "heidisql/heidisql-", CheckFile: "heidisql.exe",
			Detect: func() ([]string, map[string]string) {
				v, u := heidisql.DetectVersions()
				return []string{v}, map[string]string{v: u}
			},
			GetIcon: heidisql.GetIcon,
		},
		{
			Name: "Nginx", Category: "nginx", TargetPrefix: "nginx/nginx-", CheckFile: "nginx.exe",
			Detect: nginx.DetectVersions, GetIcon: nginx.GetIcon,
		},
		{
			Name: "OpenSSL", Category: "openssl", TargetPrefix: "openssl/openssl-", CheckFile: "bin/openssl.exe",
			Detect: func() ([]string, map[string]string) {
				v, u := openssl.DetectVersions()
				return []string{v}, map[string]string{v: u}
			},
			GetIcon: openssl.GetIcon,
		},
	}
}

func GetLatestKnownVersions() []DownloadTask {
	definitions := getPluginDefinitions()
	var tasks []DownloadTask
	baseDir := config.GetBaseDir()

	for _, def := range definitions {
		tasks = append(tasks, processPluginDefinition(def, baseDir))
	}

	return tasks
}

func processPluginDefinition(def pluginDefinition, baseDir string) DownloadTask {
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

	t.InstalledVers = getInstalledVersions(baseDir, def.Category, t.CheckFile)

	if handleSpecialCases(&t) {
		return t
	}

	currentPath := filepath.Join(baseDir, "bin", def.Category, "current")
	t.Modules = detectPluginModules(def, currentPath)

	if def.GetInfo != nil {
		t.Info = def.GetInfo(currentPath)
	}

	t.IsInstalled = isPluginInstalled(t, baseDir, currentPath)
	return t
}

func getInstalledVersions(baseDir, category, checkFile string) []string {
	installedMap := utils.GetInstalledVersionPaths(baseDir, category, checkFile)
	vers := make([]string, 0, len(installedMap))
	for v := range installedMap {
		vers = append(vers, v)
	}
	sort.Strings(vers)
	return vers
}

func handleSpecialCases(t *DownloadTask) bool {
	if t.Name == "HeidiSQL" {
		exePath, _ := utils.DetectHeidiSQLInstallation()
		if exePath != "" {
			t.IsInstalled = true
			if len(t.InstalledVers) == 0 {
				t.InstalledVers = []string{"System"}
			}
		}
		return true
	}

	if t.Name == "OpenSSL" {
		t.InstalledVers = nil
		t.IsInstalled = false
		if gv := openssl.DetectInstalledVersion(); gv != "" {
			t.Version = gv
			t.InstalledVers = []string{gv}
			t.IsInstalled = true
		}
		return true
	}
	return false
}

func detectPluginModules(def pluginDefinition, currentPath string) []PluginModule {
	if def.GetModules == nil {
		return nil
	}

	var modules []PluginModule
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
		modules = append(modules, PluginModule{
			Name: modDef.Name, IsInstalled: isModInstalled, Status: status, Version: version, CheckFile: modDef.CheckFile,
		})
	}
	return modules
}

func isPluginInstalled(t DownloadTask, baseDir, currentPath string) bool {
	// 1. Check if the currently active 'current' link is functional
	if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		return checkCheckFile(t.Name, resolved, t.CheckFile)
	}

	// 2. Fallback check for non-symlinked default target
	if t.Target != "" {
		targetPath := filepath.Join(baseDir, "bin", t.Target)
		return checkCheckFile(t.Name, targetPath, t.CheckFile)
	}

	return false
}

func checkCheckFile(pluginName, basePath, checkFile string) bool {
	cf := filepath.Join(basePath, checkFile)
	if pluginName == "Apache" {
		if _, err := os.Stat(cf); os.IsNotExist(err) {
			cf = filepath.Join(basePath, "Apache24", "bin", "httpd.exe")
		}
	}
	_, err := os.Stat(cf)
	return err == nil
}
