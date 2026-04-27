package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type DownloadTask struct {
	Name          string            `json:"name"`
	URL           string            `json:"url"`
	Version       string            `json:"version"`
	Versions      []string          `json:"versions"`
	VersionURLs   map[string]string `json:"versionUrls"`
	InstalledVers []string          `json:"installedVers"`
	Target        string            `json:"target"`
	CheckFile     string            `json:"checkFile"`
	IsInstalled   bool              `json:"isInstalled"`
	IconSVG       string            `json:"iconSvg"`
}

type Manager struct {
	ctx       context.Context
	cancels   map[string]context.CancelFunc
	cancelsMu sync.Mutex
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:     ctx,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (m *Manager) CancelDownload(name string) {
	m.cancelsMu.Lock()
	defer m.cancelsMu.Unlock()
	if cancel, ok := m.cancels[name]; ok {
		cancel()
		delete(m.cancels, name)
	}
}

func GetLatestKnownVersions() []DownloadTask {
	phpVers, phpBase := php.DetectVersions()
	apacheVers, apacheURLs := apache.DetectVersions()
	mysqlVers, mysqlURLs := mysql.DetectVersions()
	nodeVers, nodeURLs := nodejs.DetectVersions()
	pythonVers, pythonURLs := python.DetectVersions()

	nginxVersion := "1.24.0"
	nginxURL := fmt.Sprintf("https://nginx.org/download/nginx-%s.zip", nginxVersion)

	opensslVersion := "4.0.0"
	arch := utils.GetSystemArch()
	var opensslURL string
	if arch == "x64" {
		opensslURL = "https://slproweb.com/download/Win64OpenSSL_Light-4_0_0.exe"
	} else {
		opensslURL = "https://slproweb.com/download/Win32OpenSSL_Light-4_0_0.exe"
	}

	tasks := []DownloadTask{
		{
			Name: "PHP", URL: phpBase + fmt.Sprintf("php-%s-Win32-vs16-%s.zip", phpVers[0], arch),
			Version: phpVers[0], Versions: phpVers, Target: "php/php-" + phpVers[0], CheckFile: "php.exe",
			IconSVG: php.GetIcon(),
		},
		{
			Name: "Apache", URL: apacheURLs[apacheVers[0]],
			Version: apacheVers[0], Versions: apacheVers, VersionURLs: apacheURLs, Target: "apache/httpd-" + apacheVers[0], CheckFile: "bin/httpd.exe",
			IconSVG: apache.GetIcon(),
		},
		{
			Name: "MySQL", URL: mysqlURLs[mysqlVers[0]],
			Version: mysqlVers[0], Versions: mysqlVers, VersionURLs: mysqlURLs, Target: "mysql/mysql-" + mysqlVers[0], CheckFile: "bin/mysqld.exe",
			IconSVG: mysql.GetIcon(),
		},
		{
			Name: "Node.js", URL: nodeURLs[nodeVers[0]],
			Version: nodeVers[0], Versions: nodeVers, VersionURLs: nodeURLs, Target: "nodejs/node-" + nodeVers[0], CheckFile: "node.exe",
			IconSVG: nodejs.GetIcon(),
		},
		{
			Name: "Python", URL: pythonURLs[pythonVers[0]],
			Version: pythonVers[0], Versions: pythonVers, VersionURLs: pythonURLs, Target: "python/python-" + pythonVers[0], CheckFile: "python.exe",
			IconSVG: python.GetIcon(),
		},
		{
			Name: "HeidiSQL", URL: "https://www.heidisql.com/downloads/releases/HeidiSQL_12.7_64_Portable.zip",
			Version: "12.7", Target: "heidisql", CheckFile: "heidisql.exe",
			IconSVG: heidisql.GetIcon(),
		},
		{
			Name: "Nginx", URL: nginxURL,
			Version: nginxVersion, Target: "nginx/nginx-" + nginxVersion, CheckFile: "nginx.exe",
			IconSVG: nginx.GetIcon(),
		},
		{
			Name: "OpenSSL", URL: opensslURL,
			Version: opensslVersion, Target: "openssl/openssl-" + opensslVersion, CheckFile: "bin/openssl.exe",
			IconSVG: openssl.GetIcon(),
		},
	}

	baseDir := config.GetBaseDir()
	for i := range tasks {
		t := &tasks[i]
		if t.Target == "" { continue }

		category := strings.Split(filepath.ToSlash(t.Target), "/")[0]
		compDir := filepath.Join(baseDir, "bin", category)
		installedPaths := utils.GetInstalledVersionPaths(baseDir, category, t.CheckFile)
		for ver := range installedPaths { t.InstalledVers = append(t.InstalledVers, ver) }

		if t.Name == "OpenSSL" && len(t.InstalledVers) == 0 {
			if globalVer := utils.GetOpenSSLVersion("openssl"); globalVer != "" { t.InstalledVers = append(t.InstalledVers, globalVer); t.Version = globalVer }
		}

		if len(t.InstalledVers) == 0 {
			checkPath := filepath.Join(baseDir, "bin", t.Target, t.CheckFile)
			if _, err := os.Stat(checkPath); err == nil { t.InstalledVers = append(t.InstalledVers, t.Version) }
		}

		sort.Strings(t.InstalledVers)
		currentLinkPath := filepath.Join(compDir, "current")
		if resolved, err := filepath.EvalSymlinks(currentLinkPath); err == nil {
			if _, err := os.Stat(filepath.Join(resolved, t.CheckFile)); err == nil { t.IsInstalled = true } else if category == "apache" { if _, err := os.Stat(filepath.Join(resolved, "Apache24", t.CheckFile)); err == nil { t.IsInstalled = true } }
		} else {
			if _, err := os.Stat(filepath.Join(baseDir, "bin", t.Target, t.CheckFile)); err == nil { t.IsInstalled = true }
			if t.Name == "OpenSSL" && len(t.InstalledVers) > 0 { t.IsInstalled = true }
		}
	}
	return tasks
}

func (m *Manager) DeleteVersion(taskName, version string) error {
	if runtime.GOOS == "windows" {
		exeMap := map[string]string{"apache": "httpd.exe", "mysql": "mysqld.exe", "php": "php.exe", "heidisql": "heidisql.exe", "nginx": "nginx.exe", "openssl": "openssl.exe", "node.js": "node.exe", "python": "python.exe"}
		if exe := exeMap[strings.ToLower(taskName)]; exe != "" {
			killCmd := exec.Command("taskkill", "/F", "/IM", exe, "/T")
			killCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}; killCmd.Run()
			time.Sleep(500 * time.Millisecond)
		}
	}
	baseDir := config.GetBaseDir(); category := strings.ToLower(taskName)
	if category == "node.js" { category = "nodejs" }
	prefixMap := map[string]string{"php": "php-", "apache": "httpd-", "mysql": "mysql-", "nginx": "nginx-", "openssl": "openssl-", "heidisql": "", "nodejs": "node-v", "python": "python-"}
	targetDir := filepath.Join(baseDir, "bin", category, prefixMap[category]+version)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) { targetDir = filepath.Join(baseDir, "bin", category, version) }
	if _, err := os.Stat(targetDir); os.IsNotExist(err) && (category == "heidisql" || category == "nginx" || category == "openssl") { targetDir = filepath.Join(baseDir, "bin", category) }
	return os.RemoveAll(targetDir)
}

func (m *Manager) DownloadAndExtract(task DownloadTask) error {
	baseDir := config.GetBaseDir(); binDir := filepath.Join(baseDir, "bin"); targetDir := filepath.Join(binDir, task.Target)
	category := strings.Split(filepath.ToSlash(task.Target), "/")[0]
	installedPaths := utils.GetInstalledVersionPaths(baseDir, category, task.CheckFile)
	isAlreadyInstalled := false
	for ver := range installedPaths { if ver == task.Version { isAlreadyInstalled = true; break } }
	if isAlreadyInstalled {
		m.ensureCurrentLink(task); wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Ready"})
		return nil
	}
	taskCtx, cancel := context.WithCancel(m.ctx); m.cancelsMu.Lock(); m.cancels[task.Name] = cancel; m.cancelsMu.Unlock(); defer cancel()
	isZip := strings.HasSuffix(strings.ToLower(task.URL), ".zip"); ext := ".exe"
	if isZip { ext = ".zip" }
	tmpFile := filepath.Join(os.TempDir(), "ostenia_"+task.Name+ext)
	if err := DownloadFileWithContext(taskCtx, task.URL, tmpFile, task.Name); err != nil { return err }
	defer os.Remove(tmpFile)
	if !isZip {
		os.MkdirAll(targetDir, 0755); destFile := filepath.Join(targetDir, "installer.exe")
		if runtime.GOOS == "windows" { exec.Command("cmd", "/c", "move", "/Y", tmpFile, destFile).Run(); exec.Command("cmd", "/c", "start", "", destFile).Run() } else { os.Rename(tmpFile, destFile) }
		return m.ensureCurrentLink(task)
	}
	extractTmp := targetDir + ".tmp"; os.RemoveAll(extractTmp)
	if err := Unzip(taskCtx, tmpFile, extractTmp, task.Name); err != nil { return err }
	entries, _ := os.ReadDir(extractTmp)
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(extractTmp, entries[0].Name()); subEntries, _ := os.ReadDir(subDir)
		for _, se := range subEntries { os.Rename(filepath.Join(subDir, se.Name()), filepath.Join(extractTmp, se.Name())) }
		os.Remove(subDir)
	}
	os.RemoveAll(targetDir); os.Rename(extractTmp, targetDir); return m.ensureCurrentLink(task)
}

func (m *Manager) ensureCurrentLink(task DownloadTask) error {
	baseDir := config.GetBaseDir(); parts := strings.Split(filepath.ToSlash(task.Target), "/")
	if len(parts) < 2 { return nil }
	category := parts[0]; currentLink := filepath.Join(baseDir, "bin", category, "current"); targetAbs := filepath.Join(baseDir, "bin", category, parts[1])
	os.Remove(currentLink)
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "mklink", "/J", currentLink, targetAbs)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}; cmd.Run()
	} else { os.Symlink(targetAbs, currentLink) }
	return nil
}
