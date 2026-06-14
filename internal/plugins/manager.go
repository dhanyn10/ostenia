// Package plugins provides functionality for managing external components.
// This file contains the Manager struct which coordinates installation,
// deletion, and environment linking for all plugins.
package plugins

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
			c := exec.Command("taskkill", "/F", "/IM", exe, "/T")
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
	baseDir := config.GetBaseDir()
	targetDir := filepath.Join(baseDir, "bin", task.Target)

	fmt.Printf("[Manager] Request to install %s version %s to %s\n", task.Name, task.Version, targetDir)

	// 1. Strict target folder validation
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
		return nil
	}

	// 2. Start Download
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

	if err := m.downloadFile(ctx, task.URL, tmp, task.Name); err != nil {
		fmt.Printf("[Manager] Download failed: %v\n", err)
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Error: " + err.Error()})
		return err
	}

	// 3. Extraction or execution
	if !isZip && !isNupkg {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}
		dest := filepath.Join(targetDir, "installer.exe")
		// Correct copy for Windows (replace old installer if it exists)
		if err := utils.CopyFile(tmp, dest); err != nil {
			return fmt.Errorf("failed to copy installer: %w", err)
		}

		cmd := exec.Command("cmd", "/c", "start", "", dest)
		cmd.Env = utils.SafeEnv()
		utils.SetHideWindow(cmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to launch installer: %w", err)
		}

		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Completed"})
		return nil
	}

	fmt.Printf("[Manager] Extracting %s to %s\n", task.Name, targetDir)
	m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 99, Status: "Extracting..."})

	extractTmp := targetDir + ".tmp"
	_ = os.RemoveAll(extractTmp)

	if err := m.unzipFile(ctx, tmp, extractTmp, task.Name); err != nil {
		fmt.Printf("[Manager] Extraction failed: %v\n", err)
		m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Error: " + err.Error()})
		return err
	}

	// Handle Nuget package structure (files are in 'tools' folder)
	if isNupkg {
		toolsDir := filepath.Join(extractTmp, "tools")
		if _, err := os.Stat(toolsDir); err == nil {
			es, _ := os.ReadDir(toolsDir)
			for _, e := range es {
				_ = os.Rename(filepath.Join(toolsDir, e.Name()), filepath.Join(extractTmp, e.Name()))
			}
			_ = os.Remove(toolsDir)
		}
	}

	// Handle nested folders (e.g., zip containing a single root folder)
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
	if err := os.Rename(extractTmp, targetDir); err != nil {
		fmt.Printf("[Manager] Failed to rename folder: %v\n", err)
		return err
	}

	_ = os.Remove(tmp)
	_ = m.ensureCurrentLink(task)
	fmt.Printf("[Manager] %s v%s installation complete.\n", task.Name, task.Version)
	m.emit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Completed"})
	return nil
}

func (m *Manager) downloadFile(ctx context.Context, url, path, name string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	wc := &WriteCounter{
		Total:     uint64(resp.ContentLength),
		StartTime: time.Now(),
		OnProgress: func(p, t uint64, s string) {
			pct := 0.0
			if t > 0 {
				pct = (float64(p) / float64(t)) * 100
			}
			m.emit(m.ctx, "download_progress", Progress{Name: name, Percentage: pct, Status: "Downloading...", Speed: s, Downloaded: formatBytes(p)})
		},
	}
	_, err = io.Copy(out, io.TeeReader(resp.Body, wc))
	return err
}

// DownloadFileManual downloads a file without context-based cancellation
func (m *Manager) DownloadFileManual(url, path, name string) error {
	return m.downloadFile(m.ctx, url, path, name)
}

func (m *Manager) unzipFile(ctx context.Context, src, dest, name string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	total := len(r.File)
	for i, f := range r.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fpath := filepath.Join(dest, f.Name)
			if f.FileInfo().IsDir() {
				_ = os.MkdirAll(fpath, os.ModePerm)
				continue
			}
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			rc, err := f.Open()
			if err != nil {
				outFile.Close()
				return err
			}
			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return err
			}
			if i%20 == 0 { // Reduce event noise
				m.emit(m.ctx, "download_progress", Progress{Name: name, Percentage: (float64(i+1) / float64(total)) * 100, Status: "Extracting..."})
			}
		}
	}
	return nil
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
	c := exec.Command("cmd", "/c", "mklink", "/J", link, target)
	c.Env = utils.SafeEnv()
	utils.SetHideWindow(c)
	return c.Run()
}

// WriteCounter tracks the progress of a write operation
type WriteCounter struct {
	Total, Downloaded uint64
	StartTime         time.Time
	OnProgress        func(uint64, uint64, string)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += uint64(n)
	el := time.Since(wc.StartTime).Seconds()
	sp := "0 KB/s"
	if el > 0 {
		s := float64(wc.Downloaded) / el
		if s > 1024*1024 {
			sp = fmt.Sprintf("%.2f MB/s", s/(1024*1024))
		} else {
			sp = fmt.Sprintf("%.2f KB/s", s/1024)
		}
	}
	wc.OnProgress(wc.Downloaded, wc.Total, sp)
	return n, nil
}

func formatBytes(b uint64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(1024), 0
	for n := b / 1024; n >= 1024; n /= 1024 {
		div *= 1024
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
