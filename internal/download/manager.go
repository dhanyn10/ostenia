package download

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"ostenia/internal/config"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DownloadTask struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Version     string `json:"version"`
	Target      string `json:"target"`    // relative to bin/
	CheckFile   string `json:"checkFile"` // file that must exist to verify installation
	IsInstalled bool   `json:"isInstalled"`
}

type Progress struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
	Speed      string  `json:"speed"`
	Downloaded string  `json:"downloaded"` // e.g., "15.4 MB"
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

// GetLatestKnownVersions returns a list of prerequisite downloads
func GetLatestKnownVersions() []DownloadTask {
	tasks := []DownloadTask{
		{
			Name:      "PHP",
			URL:       "https://windows.php.net/downloads/releases/php-8.3.6-Win32-vs16-x64.zip",
			Version:   "8.3.6",
			Target:    "php/php-8.3.6",
			CheckFile: "php.exe",
		},
		{
			Name:      "Apache",
			URL:       "https://www.apachelounge.com/download/VS17/binaries/httpd-2.4.59-win64-VS17.zip",
			Version:   "2.4.59",
			Target:    "apache/httpd-2.4.59",
			CheckFile: "bin/httpd.exe",
		},
		{
			Name:      "MySQL",
			URL:       "https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.37-winx64.zip",
			Version:   "8.0.37",
			Target:    "mysql/mysql-8.0.37",
			CheckFile: "bin/mysqld.exe",
		},
		{
			Name:      "HeidiSQL",
			URL:       "https://www.heidisql.com/downloads/releases/HeidiSQL_12.7_64_Portable.zip",
			Version:   "12.7",
			Target:    "heidisql",
			CheckFile: "heidisql.exe",
		},
	}

	for i := range tasks {
		baseDir := config.GetBaseDir()
		// Verification: target folder exists AND check file exists
		checkPath := filepath.Join(baseDir, "bin", tasks[i].Target, tasks[i].CheckFile)
		if _, err := os.Stat(checkPath); err == nil {
			tasks[i].IsInstalled = true
		}
	}

	return tasks
}

func (m *Manager) DownloadAndExtract(task DownloadTask) error {
	fmt.Printf("[Downloader] Starting task: %s\n", task.Name)
	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin")
	targetDir := filepath.Join(binDir, task.Target)

	// Create cancellable context for this task
	taskCtx, cancel := context.WithCancel(m.ctx)
	m.cancelsMu.Lock()
	m.cancels[task.Name] = cancel
	m.cancelsMu.Unlock()
	defer func() {
		m.cancelsMu.Lock()
		delete(m.cancels, task.Name)
		m.cancelsMu.Unlock()
		cancel()
	}()

	// Robust installed check
	checkPath := filepath.Join(targetDir, task.CheckFile)
	if _, err := os.Stat(checkPath); err == nil {
		fmt.Printf("[Downloader] %s verified at %s\n", task.Name, checkPath)
		m.ensureCurrentLink(task)
		runtime.EventsEmit(m.ctx, "download_progress", Progress{
			Name:       task.Name,
			Percentage: 100,
			Status:     "Ready",
		})
		return nil
	}

	// Verify cancellation before starting
	if taskCtx.Err() != nil {
		return taskCtx.Err()
	}

	// If directory exists but check file doesn't, it might be a failed previous run
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Printf("[Downloader] Removing incomplete installation: %s\n", targetDir)
		os.RemoveAll(targetDir)
	}

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		return err
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("ostenia_%s.zip", task.Name))
	
	// Download stage
	err = m.downloadFileWithContext(taskCtx, task.URL, tmpFile, task.Name)
	if err != nil {
		if taskCtx.Err() != nil {
			runtime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Cancelled"})
		} else {
			runtime.EventsEmit(m.ctx, "download_error", map[string]string{"name": task.Name, "error": "Download failed: " + err.Error()})
		}
		return err
	}
	defer os.Remove(tmpFile)

	if taskCtx.Err() != nil {
		return taskCtx.Err()
	}

	// Extract phase...

	// Extract to temporary folder first to ensure atomicity
	extractTmp := targetDir + ".tmp"
	os.RemoveAll(extractTmp) // Clean up previous failures
	
	err = m.unzip(taskCtx, tmpFile, extractTmp, task.Name)
	if err != nil {
		os.RemoveAll(extractTmp)
		if taskCtx.Err() != nil {
			runtime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Cancelled"})
		} else {
			runtime.EventsEmit(m.ctx, "download_error", map[string]string{"name": task.Name, "error": "Extraction failed: " + err.Error()})
		}
		return err
	}

	// Rename to final targetDir
	os.RemoveAll(targetDir)
	err = os.Rename(extractTmp, targetDir)
	if err != nil {
		fmt.Printf("[Downloader] Rename error: %v, trying manual move if cross-device\n", err)
		// Fallback for cross-device if necessary, but here it should be same disk
		return err
	}

	fmt.Printf("[Downloader] Task %s completed successfully\n", task.Name)
	return m.ensureCurrentLink(task)
}

func (m *Manager) ensureCurrentLink(task DownloadTask) error {
	baseDir := config.GetBaseDir()
	
	// Normalize target path for splitting
	normalizedTarget := filepath.ToSlash(task.Target)
	parts := strings.Split(normalizedTarget, "/")
	
	if len(parts) < 2 {
		fmt.Printf("[Downloader] Skipping symlink for %s (flat structure)\n", task.Name)
		return nil
	}

	category := parts[0] // e.g., "php"
	categoryDir := filepath.Join(baseDir, "bin", category)
	currentLink := filepath.Join(categoryDir, "current")
	targetRel := parts[1] // e.g., "php-8.3.6"
	targetAbs := filepath.Join(categoryDir, targetRel)

	fmt.Printf("[Downloader] Linking %s -> %s\n", currentLink, targetAbs)

	// Remove existing link if it exists
	if _, err := os.Lstat(currentLink); err == nil {
		os.Remove(currentLink)
	}

	// Create symlink/junction
	// On Windows, Symlink needs privilege, but Junction doesn't necessarily or we can create it via cmd.
	// For simplicity and portability, let's just use os.Symlink but log failure.
	err := os.Symlink(targetAbs, currentLink)
	if err != nil {
		// Fallback: If symlink fails (e.g. no privileges), we could try a Directory Junction (Windows)
		// But for now, we'll just log it.
		fmt.Printf("Warning: Could not create symlink for %s: %v\n", task.Name, err)
	}

	return nil
}

func (m *Manager) downloadFileWithContext(ctx context.Context, url string, filepath string, name string) error {
	out, err := os.Create(filepath)
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
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	counter := &WriteCounter{
		Total:     uint64(resp.ContentLength),
		StartTime: time.Now(),
		OnProgress: func(progress uint64, total uint64, speed string) {
			var percentage float64
			status := "Downloading..."
			if total > 0 {
				percentage = (float64(progress) / float64(total)) * 100
			} else {
				status = "Downloading (Streaming)..."
			}
			runtime.EventsEmit(m.ctx, "download_progress", Progress{
				Name:       name,
				Percentage: percentage,
				Status:     status,
				Speed:      speed,
				Downloaded: formatBytes(progress),
			})
		},
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	return err
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (m *Manager) unzip(ctx context.Context, src string, dest string, name string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	totalFiles := len(r.File)
	for i, f := range r.File {
		// Check cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		err := m.extractAndWriteFile(dest, f)
		if err != nil {
			return err
		}
		percentage := (float64(i+1) / float64(totalFiles)) * 100
		runtime.EventsEmit(m.ctx, "download_progress", Progress{
			Name:       name,
			Percentage: percentage,
			Status:     fmt.Sprintf("Extracting %d/%d...", i+1, totalFiles),
		})
	}

	runtime.EventsEmit(m.ctx, "download_progress", Progress{
		Name:       name,
		Percentage: 100,
		Status:     "Completed",
	})
	return nil
}

func (m *Manager) extractAndWriteFile(dest string, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dest, f.Name)

	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", path)
	}

	if f.FileInfo().IsDir() {
		os.MkdirAll(path, f.Mode())
	} else {
		os.MkdirAll(filepath.Dir(path), f.Mode())
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

type WriteCounter struct {
	Total      uint64
	Downloaded uint64
	StartTime  time.Time
	OnProgress func(progress uint64, total uint64, speed string)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += uint64(n)

	// Calculate speed
	elapsed := time.Since(wc.StartTime).Seconds()
	var speedStr string
	if elapsed > 0 {
		speed := float64(wc.Downloaded) / elapsed
		if speed > 1024*1024 {
			speedStr = fmt.Sprintf("%.2f MB/s", speed/(1024*1024))
		} else {
			speedStr = fmt.Sprintf("%.2f KB/s", speed/1024)
		}
	}

	wc.OnProgress(wc.Downloaded, wc.Total, speedStr)
	return n, nil
}
