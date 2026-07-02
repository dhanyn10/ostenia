// Package plugins provides functionality for managing external components.
// This file contains the Manager struct which coordinates installation,
// deletion, and environment linking for all plugins.
package plugins

import (
	"context"
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Manager coordinates the download, extraction, and management of plugin versions
type Manager struct {
	ctx       context.Context
	cancels   map[string]context.CancelFunc
	cancelsMu sync.Mutex
	emit      func(ctx context.Context, eventName string, optionalData ...interface{})
}

func (m *Manager) GetInstalledVersionPaths(category, checkFile string) map[string]string {
	return utils.GetInstalledVersionPaths(config.GetBaseDir(), category, checkFile)
}

func (m *Manager) InstallModule(moduleName string, phpPath string, emitProgress func(string, float64, string)) error {
	// Actual implementation is in subpackages, but Manager provides the entry point for the interface
	// This is a bit tricky because the interface expects it here.
	// For now, the App calls subpackages directly, so we just satisfy the interface.
	return nil
}

func (m *Manager) UninstallModule(moduleName string, phpPath string) error {
	return nil
}

// NewManager creates a new plugin Manager instance
func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:     ctx,
		cancels: make(map[string]context.CancelFunc),
		emit:    wruntime.EventsEmit,
	}
}

// CancelDownload aborts an ongoing download task
func (m *Manager) CancelDownload(name string) {
	m.cancelsMu.Lock()
	defer m.cancelsMu.Unlock()
	if c, ok := m.cancels[name]; ok {
		c()
		delete(m.cancels, name)
	}
}

// DeleteVersion removes a specific version directory and kills any related processes on Windows
func (m *Manager) DeleteVersion(taskName, version string) error {
	if runtime.GOOS == "windows" {
		exeMap := map[string]string{
			"apache": "httpd.exe", "mysql": "mysqld.exe", "php": "php.exe",
			"heidisql": "heidisql.exe", "nginx": "nginx.exe", "openssl": "openssl.exe",
			"node.js": "node.exe", "python": "python.exe",
		}
		if exe, ok := exeMap[strings.ToLower(taskName)]; ok {
			taskkillPath := filepath.Join(utils.GetSystemDirectory(), "taskkill.exe")
			c := utils.Executor.Command(taskkillPath, "/F", "/IM", exe, "/T")
			c.Env = utils.SafeEnv()
			utils.SetHideWindow(c)
			_ = c.Run()
			time.Sleep(500 * time.Millisecond)
		}
	}
	baseDir := config.GetBaseDir()
	cat := strings.ToLower(taskName)
	if cat == "node.js" {
		cat = "nodejs"
	}
	pMap := map[string]string{
		"php": "php-", "apache": "httpd-", "mysql": "mysql-",
		"nginx": "nginx-", "openssl": "openssl-", "nodejs": "node-v",
		"python": "python-",
	}
	dir := filepath.Join(baseDir, "bin", cat, pMap[cat]+version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = filepath.Join(baseDir, "bin", cat, version)
	}
	return os.RemoveAll(dir)
}

// DownloadAndExtract handles the full lifecycle of installing a plugin task
func (m *Manager) DownloadAndExtract(task DownloadTask) error {
	targetDir := filepath.Join(config.GetBaseDir(), "bin", task.Target)
	fmt.Printf("[Manager] Request to install %s version %s to %s\n", task.Name, task.Version, targetDir)

	if m.isAlreadyInstalled(task, targetDir) {
		return nil
	}

	tmpFile, isArchive, err := m.downloadPlugin(task)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	if !isArchive {
		return m.handleInstaller(task, tmpFile, targetDir)
	}

	return m.handleArchive(task, tmpFile, targetDir)
}

func (m *Manager) isAlreadyInstalled(task DownloadTask, targetDir string) bool {
	checkFile := filepath.Join(targetDir, task.CheckFile)
	if task.Name == "Apache" {
		if _, err := os.Stat(checkFile); os.IsNotExist(err) {
			checkFile = filepath.Join(targetDir, "Apache24", "bin", "httpd.exe")
		}
	}

	if _, err := os.Stat(checkFile); err == nil {
		fmt.Printf("[Manager] %s v%s already exists. Linking only.\n", task.Name, task.Version)
		_ = m.ensureCurrentLink(task)
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Ready"})
		return true
	}
	return false
}

func (m *Manager) downloadPlugin(task DownloadTask) (string, bool, error) {
	fmt.Printf("[Manager] Downloading %s from %s\n", task.Name, task.URL)
	ctx, cancel := context.WithCancel(m.ctx)
	m.cancelsMu.Lock()
	m.cancels[task.Name] = cancel
	m.cancelsMu.Unlock()
	defer func() {
		cancel()
		m.cancelsMu.Lock()
		delete(m.cancels, task.Name)
		m.cancelsMu.Unlock()
	}()

	isZip := strings.HasSuffix(strings.ToLower(task.URL), ".zip")
	isNupkg := strings.HasSuffix(strings.ToLower(task.URL), ".nupkg")
	ext := ".exe"
	if isZip {
		ext = ".zip"
	} else if isNupkg {
		ext = ".nupkg"
	}
	tmp := filepath.Join(os.TempDir(), "ostenia_"+task.Name+ext)

	err := utils.DownloadFile(ctx, task.URL, tmp, task.Name, func(pct float64, status, speed, downloaded string) {
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: pct, Status: status, Speed: speed, Downloaded: downloaded})
	})

	if err != nil {
		fmt.Printf("[Manager] Download failed: %v\n", err)
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Error: " + err.Error()})
		return "", false, err
	}

	return tmp, isZip || isNupkg, nil
}

func (m *Manager) handleInstaller(task DownloadTask, tmpFile, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	dest := filepath.Join(targetDir, "installer.exe")
	if err := utils.CopyFile(tmpFile, dest); err != nil {
		return fmt.Errorf("failed to copy installer: %w", err)
	}

	cmdPath := filepath.Join(utils.GetSystemDirectory(), "cmd.exe")
	cmd := utils.Executor.Command(cmdPath, "/c", "start", "", dest)
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Completed"})
	return nil
}

func (m *Manager) handleArchive(task DownloadTask, tmpFile, targetDir string) error {
	fmt.Printf("[Manager] Extracting %s to %s\n", task.Name, targetDir)
	m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 99, Status: "Extracting..."})

	extractTmp := targetDir + ".tmp"
	_ = os.RemoveAll(extractTmp)

	if err := Unzip(m.ctx, tmpFile, extractTmp, task.Name, m.emit); err != nil {
		fmt.Printf("[Manager] Extraction failed: %v\n", err)
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Error: " + err.Error()})
		return err
	}

	if err := m.postProcessExtraction(task, extractTmp, targetDir); err != nil {
		fmt.Printf("[Manager] Post-processing failed: %v\n", err)
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Error: " + err.Error()})
		return err
	}

	_ = m.ensureCurrentLink(task)
	fmt.Printf("[Manager] %s v%s installation complete.\n", task.Name, task.Version)
	m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Completed"})
	return nil
}

func (m *Manager) postProcessExtraction(task DownloadTask, extractTmp, targetDir string) error {
	// Handle Nuget package structure
	if strings.HasSuffix(strings.ToLower(task.URL), ".nupkg") {
		toolsDir := filepath.Join(extractTmp, "tools")
		if _, err := os.Stat(toolsDir); err == nil {
			es, _ := os.ReadDir(toolsDir)
			for _, e := range es {
				_ = os.Rename(filepath.Join(toolsDir, e.Name()), filepath.Join(extractTmp, e.Name()))
			}
			_ = os.Remove(toolsDir)
		}
	}

	// Handle nested folders
	es, _ := os.ReadDir(extractTmp)
	if len(es) == 1 && es[0].IsDir() {
		sDir := filepath.Join(extractTmp, es[0].Name())
		ses, _ := os.ReadDir(sDir)
		for _, se := range ses {
			_ = os.Rename(filepath.Join(sDir, se.Name()), filepath.Join(extractTmp, se.Name()))
		}
		_ = os.Remove(sDir)
	}

	_ = os.RemoveAll(targetDir)
	return os.Rename(extractTmp, targetDir)
}

// DownloadFileManual downloads a file without context-based cancellation
func (m *Manager) DownloadFileManual(url, path, name string) error {
	return utils.DownloadFile(m.ctx, url, path, name, func(pct float64, status, speed, downloaded string) {
		m.emit(m.ctx, "download_progress", Progress{Name: name, Percentage: pct, Status: status, Speed: speed, Downloaded: downloaded})
	})
}


func (m *Manager) ensureCurrentLink(task DownloadTask) error {
	baseDir := config.GetBaseDir()
	parts := strings.Split(filepath.ToSlash(task.Target), "/")
	if len(parts) < 2 {
		return nil
	}
	link := filepath.Join(baseDir, "bin", parts[0], "current")
	target := filepath.Join(baseDir, "bin", task.Target)
	_ = os.Remove(link) // Remove old junction
	cmdPath := filepath.Join(utils.GetSystemDirectory(), "cmd.exe")
	c := utils.Executor.Command(cmdPath, "/c", "mklink", "/J", link, target)
	c.Env = utils.SafeEnv()
	utils.SetHideWindow(c)
	return c.Run()
}

