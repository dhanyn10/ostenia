package download

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
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
}

type Progress struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
	Speed      string  `json:"speed"`
	Downloaded string  `json:"downloaded"`
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
	phpVers, phpBase := DetectPHPVersions()
	apacheVers, apacheURLs := DetectApacheVersions()
	mysqlVers, mysqlURLs := DetectMySQLVersions()

	tasks := []DownloadTask{
		{
			Name:     "PHP",
			URL:      phpBase + fmt.Sprintf("php-%s-Win32-vs16-%s.zip", phpVers[0], getSystemArch()),
			Version:  phpVers[0],
			Versions: phpVers,
			Target:   "php/php-" + phpVers[0],
			CheckFile: "php.exe",
		},
		{
			Name:        "Apache",
			URL:         apacheURLs[apacheVers[0]],
			Version:     apacheVers[0],
			Versions:    apacheVers,
			VersionURLs: apacheURLs,
			Target:      "apache/httpd-" + apacheVers[0],
			CheckFile:   "bin/httpd.exe",
		},
		{
			Name:        "MySQL",
			URL:         mysqlURLs[mysqlVers[0]],
			Version:     mysqlVers[0],
			Versions:    mysqlVers,
			VersionURLs: mysqlURLs,
			Target:      "mysql/mysql-" + mysqlVers[0],
			CheckFile:   "bin/mysqld.exe",
		},
		{
			Name:      "HeidiSQL",
			URL:       "https://www.heidisql.com/downloads/releases/HeidiSQL_12.7_64_Portable.zip",
			Version:   "12.7",
			Target:    "heidisql",
			CheckFile: "heidisql.exe",
		},
	}

	baseDir := config.GetBaseDir()
	for i := range tasks {
		t := &tasks[i]
		category := strings.Split(filepath.ToSlash(t.Target), "/")[0]
		compDir := filepath.Join(baseDir, "bin", category)
		
		entries, _ := os.ReadDir(compDir)
		for _, e := range entries {
			if e.IsDir() && e.Name() != "current" {
				ver := e.Name()
				if idx := strings.Index(ver, "-"); idx != -1 { ver = ver[idx+1:] }

				checkPaths := []string{
					filepath.Join(compDir, e.Name(), t.CheckFile),
					filepath.Join(compDir, e.Name(), "Apache24", t.CheckFile),
				}
				for _, p := range checkPaths {
					if _, err := os.Stat(p); err == nil {
						t.InstalledVers = append(t.InstalledVers, ver)
						break
					}
				}
			}
		}

		// Special case for flat structures like HeidiSQL which don't use version subfolders
		if len(t.InstalledVers) == 0 {
			checkPath := filepath.Join(baseDir, "bin", t.Target, t.CheckFile)
			if _, err := os.Stat(checkPath); err == nil {
				t.InstalledVers = append(t.InstalledVers, t.Version)
				t.IsInstalled = true
			}
		}

		sort.Strings(t.InstalledVers)

		currentPath := filepath.Join(compDir, "current")
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
			if _, err := os.Stat(filepath.Join(resolved, t.CheckFile)); err == nil {
				t.IsInstalled = true
			} else if _, err := os.Stat(filepath.Join(resolved, "Apache24", t.CheckFile)); err == nil {
				t.IsInstalled = true
			}
		} else {
			// If no 'current' symlink, check if the direct target exists (for flat structures)
			if _, err := os.Stat(filepath.Join(baseDir, "bin", t.Target, t.CheckFile)); err == nil {
				t.IsInstalled = true
			}
		}
	}
	return tasks
}

func (m *Manager) DeleteVersion(taskName, version string) error {
	if runtime.GOOS == "windows" {
		exeMap := map[string]string{"apache": "httpd.exe", "mysql": "mysqld.exe", "php": "php.exe", "heidisql": "heidisql.exe"}
		if exe := exeMap[strings.ToLower(taskName)]; exe != "" {
			exec.Command("taskkill", "/F", "/IM", exe, "/T").Run()
			time.Sleep(500 * time.Millisecond)
		}
	}

	baseDir := config.GetBaseDir()
	category := strings.ToLower(taskName)
	prefixMap := map[string]string{"php": "php-", "apache": "httpd-", "mysql": "mysql-", "heidisql": ""}

	targetDir := filepath.Join(baseDir, "bin", category, prefixMap[category]+version)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		targetDir = filepath.Join(baseDir, "bin", category, version)
	}

	// Final fallback for flat structure
	if _, err := os.Stat(targetDir); os.IsNotExist(err) && category == "heidisql" {
		targetDir = filepath.Join(baseDir, "bin", category)
	}

	return os.RemoveAll(targetDir)
}

func (m *Manager) DownloadAndExtract(task DownloadTask) error {
	baseDir := config.GetBaseDir()
	targetDir := filepath.Join(baseDir, "bin", task.Target)

	if _, err := os.Stat(filepath.Join(targetDir, task.CheckFile)); err == nil {
		m.ensureCurrentLink(task)
		wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Ready"})
		return nil
	}

	taskCtx, cancel := context.WithCancel(m.ctx)
	m.cancelsMu.Lock()
	m.cancels[task.Name] = cancel
	m.cancelsMu.Unlock()
	defer cancel()

	tmpFile := filepath.Join(os.TempDir(), "ostenia_"+task.Name+".zip")
	if err := m.downloadFileWithContext(taskCtx, task.URL, tmpFile, task.Name); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	extractTmp := targetDir + ".tmp"
	os.RemoveAll(extractTmp)
	if err := m.unzip(taskCtx, tmpFile, extractTmp, task.Name); err != nil {
		return err
	}

	if entries, _ := os.ReadDir(extractTmp); len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(extractTmp, entries[0].Name())
		subEntries, _ := os.ReadDir(subDir)
		for _, se := range subEntries {
			os.Rename(filepath.Join(subDir, se.Name()), filepath.Join(extractTmp, se.Name()))
		}
		os.Remove(subDir)
	}

	os.RemoveAll(targetDir)
	os.Rename(extractTmp, targetDir)
	return m.ensureCurrentLink(task)
}

func (m *Manager) ensureCurrentLink(task DownloadTask) error {
	baseDir := config.GetBaseDir()
	parts := strings.Split(filepath.ToSlash(task.Target), "/")
	if len(parts) < 2 { return nil }

	currentLink := filepath.Join(baseDir, "bin", parts[0], "current")
	targetAbs := filepath.Join(baseDir, "bin", parts[0], parts[1])

	os.Remove(currentLink)
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", "mklink", "/J", currentLink, targetAbs).Run()
	}
	return os.Symlink(targetAbs, currentLink)
}

func (m *Manager) downloadFileWithContext(ctx context.Context, url, filepath, name string) error {
	out, _ := os.Create(filepath)
	defer out.Close()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	counter := &WriteCounter{Total: uint64(resp.ContentLength), StartTime: time.Now(), OnProgress: func(p, t uint64, s string) {
		pct := 0.0
		if t > 0 { pct = (float64(p) / float64(t)) * 100 }
		wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: name, Percentage: pct, Status: "Downloading...", Speed: s, Downloaded: formatBytes(p)})
	}}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	return err
}

func formatBytes(b uint64) string {
	units := []string{"B", "KB", "MB", "GB"}
	val := float64(b)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", val, units[i])
}

func (m *Manager) unzip(ctx context.Context, src, dest, name string) error {
	r, err := zip.OpenReader(src)
	if err != nil { return err }
	defer r.Close()
	os.MkdirAll(dest, 0755)

	for i, f := range r.File {
		if ctx.Err() != nil { return ctx.Err() }
		m.extractFile(dest, f)
		wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: name, Percentage: (float64(i+1)/float64(len(r.File)))*100, Status: fmt.Sprintf("Extracting %d/%d...", i+1, len(r.File))})
	}
	return nil
}

func (m *Manager) extractFile(dest string, f *zip.File) {
	rc, _ := f.Open()
	defer rc.Close()
	path := filepath.Join(dest, f.Name)
	if f.FileInfo().IsDir() {
		os.MkdirAll(path, f.Mode())
	} else {
		os.MkdirAll(filepath.Dir(path), f.Mode())
		out, _ := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		defer out.Close()
		io.Copy(out, rc)
	}
}

type WriteCounter struct {
	Total, Downloaded uint64
	StartTime time.Time
	OnProgress func(p, t uint64, s string)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += uint64(n)
	elapsed := time.Since(wc.StartTime).Seconds()
	speed := ""
	if elapsed > 0 {
		s := float64(wc.Downloaded) / elapsed
		if s > 1024*1024 { speed = fmt.Sprintf("%.2f MB/s", s/(1024*1024)) } else { speed = fmt.Sprintf("%.2f KB/s", s/1024) }
	}
	wc.OnProgress(wc.Downloaded, wc.Total, speed)
	return n, nil
}
