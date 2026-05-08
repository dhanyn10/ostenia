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
	Name      string
	Category string
	TargetPrefix string
	CheckFile string
	Detect    func() ([]string, map[string]string)
	GetIcon   func() string
	GetModules func() []utils.ModuleDefinition
	GetModuleVersion func(name string, path string) string
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

	var tasks []DownloadTask
	baseDir := config.GetBaseDir()

	for _, def := range definitions {
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

		// Special case for OpenSSL
		if t.Name == "OpenSSL" {
			t.InstalledVers = nil
			t.IsInstalled = false
			if gv := openssl.DetectInstalledVersion(); gv != "" {
				t.Version = gv
				t.InstalledVers = []string{gv}
				t.IsInstalled = true
			}
			tasks = append(tasks, t)
			continue
		}

		// 1.5 Detect Modules (Generically)
		currentPath := filepath.Join(baseDir, "bin", def.Category, "current")
		if def.GetModules != nil {
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

		// 2. Check if the currently active 'current' link is functional
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
			cf := filepath.Join(resolved, t.CheckFile)
			if t.Name == "Apache" {
				if _, err := os.Stat(cf); os.IsNotExist(err) {
					cf = filepath.Join(resolved, "Apache24", "bin", "httpd.exe")
				}
			}
			if _, err := os.Stat(cf); err == nil {
				t.IsInstalled = true
			}
		} else {
			// Fallback check for non-symlinked default target
			if t.Target != "" {
				cf := filepath.Join(baseDir, "bin", t.Target, t.CheckFile)
				if t.Name == "Apache" {
					if _, err := os.Stat(cf); os.IsNotExist(err) {
						cf = filepath.Join(baseDir, "bin", t.Target, "Apache24", "bin", "httpd.exe")
					}
				}
				if _, err := os.Stat(cf); err == nil {
					t.IsInstalled = true
				}
			}
		}

		tasks = append(tasks, t)
	}

	return tasks
}
