// Package plugins provides functionality for managing external components.
// This file focuses on discovering available versions from the internet
// and detecting versions already installed on the local system.
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
	"strings"
)

// GetLatestKnownVersions aggregates available and installed plugin versions.
// It scans both remote repositories and local installation directories.
func GetLatestKnownVersions() []DownloadTask {
	phpVers, phpUrls := php.DetectVersions()
	apacheVers, apacheUrls := apache.DetectVersions()
	mysqlVers, mysqlUrls := mysql.DetectVersions()
	nodeVers, nodeUrls := nodejs.DetectVersions()
	pythonVers, pythonUrls := python.DetectVersions()
	opensslVer, opensslUrl := openssl.DetectVersions()
	nginxVers, nginxUrls := nginx.DetectVersions()
	heidiVer, heidiUrl := heidisql.DetectVersions()

	tasks := []DownloadTask{
		{Name: "PHP", URL: phpUrls[phpVers[0]], Version: phpVers[0], Versions: phpVers, VersionUrls: phpUrls, Target: "php/php-" + phpVers[0], CheckFile: "php.exe", IconSVG: php.GetIcon()},
		{Name: "Apache", URL: apacheUrls[apacheVers[0]], Version: apacheVers[0], Versions: apacheVers, VersionUrls: apacheUrls, Target: "apache/httpd-" + apacheVers[0], CheckFile: "bin/httpd.exe", IconSVG: apache.GetIcon()},
		{Name: "MySQL", URL: mysqlUrls[mysqlVers[0]], Version: mysqlVers[0], Versions: mysqlVers, VersionUrls: mysqlUrls, Target: "mysql/mysql-" + mysqlVers[0], CheckFile: "bin/mysqld.exe", IconSVG: mysql.GetIcon()},
		{Name: "Node.js", URL: nodeUrls[nodeVers[0]], Version: nodeVers[0], Versions: nodeVers, VersionUrls: nodeUrls, Target: "nodejs/node-v" + nodeVers[0], CheckFile: "node.exe", IconSVG: nodejs.GetIcon()},
		{Name: "Python", URL: pythonUrls[pythonVers[0]], Version: pythonVers[0], Versions: pythonVers, VersionUrls: pythonUrls, Target: "python/python-" + pythonVers[0], CheckFile: "python.exe", IconSVG: python.GetIcon()},
		{Name: "HeidiSQL", URL: heidiUrl, Version: heidiVer, Versions: []string{heidiVer}, Target: "heidisql/heidisql-" + heidiVer, CheckFile: "heidisql.exe", IconSVG: heidisql.GetIcon()},
		{Name: "Nginx", URL: nginxUrls[nginxVers[0]], Version: nginxVers[0], Versions: nginxVers, VersionUrls: nginxUrls, Target: "nginx/nginx-" + nginxVers[0], CheckFile: "nginx.exe", IconSVG: nginx.GetIcon()},
		{Name: "OpenSSL", URL: opensslUrl, Version: opensslVer, Target: "openssl/openssl-" + opensslVer, CheckFile: "bin/openssl.exe", IconSVG: openssl.GetIcon()},
	}

	baseDir := config.GetBaseDir()
	for i := range tasks {
		t := &tasks[i]
		category := strings.Split(filepath.ToSlash(t.Target), "/")[0]

		// 1. Detect ALL installed versions
		installedMap := utils.GetInstalledVersionPaths(baseDir, category, t.CheckFile)
		t.InstalledVers = make([]string, 0, len(installedMap))
		for v := range installedMap {
			t.InstalledVers = append(t.InstalledVers, v)
		}

		if t.Name == "OpenSSL" {
			t.InstalledVers = nil
			t.IsInstalled = false
			if gv := utils.GetOpenSSLVersion("openssl"); gv != "" {
				t.Version = gv
				t.InstalledVers = []string{gv}
				t.IsInstalled = true
			}
		}
		sort.Strings(t.InstalledVers)
		if t.Name == "OpenSSL" {
			continue
		}

		// 2. Check if the currently active 'current' link is functional
		currentPath := filepath.Join(baseDir, "bin", category, "current")
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
	return tasks
}
