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
	"regexp"
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

// getInstalledVersionPaths checks for installed versions and returns their full paths.
func getInstalledVersionPaths(baseDir string, category string, checkFile string) map[string]string {
	installedPaths := make(map[string]string)
	compDir := filepath.Join(baseDir, "bin", category)
	entries, err := os.ReadDir(compDir)
	if err != nil {
		return installedPaths
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			ver := entry.Name()
			if idx := strings.Index(ver, "-"); idx != -1 {
				ver = ver[idx+1:]
			}

			potentialPaths := []string{
				filepath.Join(compDir, entry.Name(), checkFile),
				filepath.Join(compDir, entry.Name(), "Apache24", checkFile),
			}

			for _, p := range potentialPaths {
				if _, err := os.Stat(p); err == nil {
					installedPaths[ver] = p
					break
				}
			}
		}
	}
	return installedPaths
}

// getOpenSSLVersion executes 'openssl version' and parses the output.
// It can take an absolute path to the exe or just "openssl" for global check.
func getOpenSSLVersion(opensslCmd string) string {
	cmd := exec.Command(opensslCmd, "version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Example output: OpenSSL 3.2.1 1 Feb 2024 (Library: OpenSSL 3.2.1 1 Feb 2024)
	re := regexp.MustCompile(`OpenSSL\s+([\d\.]+[a-z]?)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return "Installed" // Fallback if parsing fails but command works
}

func GetLatestKnownVersions() []DownloadTask {
	phpVers, phpBase := DetectPHPVersions()
	apacheVers, apacheURLs := DetectApacheVersions()
	mysqlVers, mysqlURLs := DetectMySQLVersions()

	nginxVersion := "1.24.0"
	nginxURL := fmt.Sprintf("https://nginx.org/download/nginx-%s.zip", nginxVersion)

	opensslVersion := "4.0.0"
	arch := getSystemArch()
	var opensslURL string
	if arch == "x64" {
		opensslURL = "https://slproweb.com/download/Win64OpenSSL_Light-4_0_0.exe"
	} else {
		opensslURL = "https://slproweb.com/download/Win32OpenSSL_Light-4_0_0.exe"
	}

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
		{
			Name:      "Nginx",
			URL:       nginxURL,
			Version:   nginxVersion,
			Target:    "nginx/nginx-" + nginxVersion,
			CheckFile: "nginx.exe",
		},
		{
			Name:      "OpenSSL",
			URL:       opensslURL,
			Version:   opensslVersion,
			Target:    "openssl/openssl-" + opensslVersion,
			CheckFile: "bin/openssl.exe",
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
					filepath.Join(compDir, e.Name(), "bin", t.CheckFile),
				}
				for _, p := range checkPaths {
					if _, err := os.Stat(p); err == nil {
						if t.Name == "OpenSSL" {
							detectedVer := getOpenSSLVersion(p)
							if detectedVer != "" {
								t.InstalledVers = append(t.InstalledVers, detectedVer)
							} else {
								t.InstalledVers = append(t.InstalledVers, ver)
							}
						} else {
							t.InstalledVers = append(t.InstalledVers, ver)
						}
						break
					}
				}
			}
		}

		// Local check for flat structures
		if len(t.InstalledVers) == 0 {
			checkPath := filepath.Join(baseDir, "bin", t.Target, t.CheckFile)
			if _, err := os.Stat(checkPath); err == nil {
				if t.Name == "OpenSSL" {
					detectedVer := getOpenSSLVersion(checkPath)
					if detectedVer != "" {
						t.InstalledVers = append(t.InstalledVers, detectedVer)
					} else {
						t.InstalledVers = append(t.InstalledVers, t.Version)
					}
				} else {
					t.InstalledVers = append(t.InstalledVers, t.Version)
				}
			}
		}

		// Global check for OpenSSL
		if t.Name == "OpenSSL" && len(t.InstalledVers) == 0 {
			globalVer := getOpenSSLVersion("openssl")
			if globalVer != "" {
				t.InstalledVers = append(t.InstalledVers, globalVer)
				// We also set the version to match global if needed to trigger "Ready"
				t.Version = globalVer
			}
		}

		sort.Strings(t.InstalledVers)

		currentPath := filepath.Join(compDir, "current")
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
			if _, err := os.Stat(filepath.Join(resolved, t.CheckFile)); err == nil {
				t.IsInstalled = true
			} else if category == "apache" {
				if _, err := os.Stat(filepath.Join(resolved, "Apache24", t.CheckFile)); err == nil {
					t.IsInstalled = true
				}
			} else if category == "openssl" {
				if _, err := os.Stat(filepath.Join(resolved, "bin", t.CheckFile)); err == nil {
					t.IsInstalled = true
				}
			}
		} else {
			if _, err := os.Stat(filepath.Join(baseDir, "bin", t.Target, t.CheckFile)); err == nil {
				t.IsInstalled = true
			}
			// If globally installed, mark as installed
			if t.Name == "OpenSSL" && len(t.InstalledVers) > 0 {
				t.IsInstalled = true
			}
		}
	}
	return tasks
}

func (m *Manager) DeleteVersion(taskName, version string) error {
	if runtime.GOOS == "windows" {
		exeMap := map[string]string{"apache": "httpd.exe", "mysql": "mysqld.exe", "php": "php.exe", "heidisql": "heidisql.exe", "nginx": "nginx.exe", "openssl": "openssl.exe"}
		if exe := exeMap[strings.ToLower(taskName)]; exe != "" {
			exec.Command("taskkill", "/F", "/IM", exe, "/T").Run()
			time.Sleep(500 * time.Millisecond)
		}
	}

	baseDir := config.GetBaseDir()
	category := strings.ToLower(taskName)
	prefixMap := map[string]string{"php": "php-", "apache": "httpd-", "mysql": "mysql-", "nginx": "nginx-", "openssl": "openssl-", "heidisql": ""}

	targetDir := filepath.Join(baseDir, "bin", category, prefixMap[category]+version)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		targetDir = filepath.Join(baseDir, "bin", category, version)
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) && (category == "heidisql" || category == "nginx" || category == "openssl") {
		targetDir = filepath.Join(baseDir, "bin", category)
	}

	return os.RemoveAll(targetDir)
}

func (m *Manager) DownloadAndExtract(task DownloadTask) error {
	baseDir := config.GetBaseDir()
	targetDir := filepath.Join(baseDir, "bin", task.Target)

	category := ""
	if strings.Contains(task.Target, "/") {
		category = strings.Split(task.Target, "/")[0]
	} else {
		category = task.Target
	}
	installedPaths := getInstalledVersionPaths(baseDir, category, task.CheckFile)
	isAlreadyInstalled := false
	for ver := range installedPaths {
		if ver == task.Version {
			isAlreadyInstalled = true
			break
		}
	}

	if isAlreadyInstalled {
		fmt.Printf("[Downloader] %s verified at %s\n", task.Name, targetDir)
		m.ensureCurrentLink(task)
		wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 100, Status: "Ready"})
		return nil
	}

	taskCtx, cancel := context.WithCancel(m.ctx)
	m.cancelsMu.Lock()
	m.cancels[task.Name] = cancel
	m.cancelsMu.Unlock()
	defer cancel()

	isZip := strings.HasSuffix(strings.ToLower(task.URL), ".zip")
	ext := ".zip"
	if !isZip {
		ext = ".exe"
	}

	tmpFile := filepath.Join(os.TempDir(), "ostenia_"+task.Name+ext)
	if err := m.downloadFileWithContext(taskCtx, task.URL, tmpFile, task.Name); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	if !isZip {
		os.MkdirAll(targetDir, 0755)
		destFile := filepath.Join(targetDir, "installer.exe")

		fmt.Printf("[Downloader] Moving installer using system command: %s\n", destFile)
		if runtime.GOOS == "windows" {
			err := exec.Command("cmd", "/c", "move", "/Y", tmpFile, destFile).Run()
			if err != nil {
				return fmt.Errorf("system move failed: %w", err)
			}

			fmt.Printf("[Downloader] Launching installer: %s\n", destFile)
			exec.Command("cmd", "/c", "start", "", destFile).Run()
		} else {
			os.Rename(tmpFile, destFile)
		}

		return m.ensureCurrentLink(task)
	}

	extractTmp := targetDir + ".tmp"
	os.RemoveAll(extractTmp)
	if err := m.unzip(taskCtx, tmpFile, extractTmp, task.Name); err != nil {
		return err
	}

	entries, _ := os.ReadDir(extractTmp)
	if len(entries) == 1 && entries[0].IsDir() {
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
	if len(parts) < 2 {
		return nil
	}

	category := parts[0]
	currentLink := filepath.Join(baseDir, "bin", category, "current")
	targetAbs := filepath.Join(baseDir, "bin", category, parts[1])

	os.Remove(currentLink)
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "mklink", "/J", currentLink, targetAbs)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Warning: Could not create junction for %s: %v\n", task.Name, err)
		}
	} else {
		err := os.Symlink(targetAbs, currentLink)
		if err != nil {
			fmt.Printf("Warning: Could not create symlink for %s: %v\n", task.Name, err)
		}
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
			wruntime.EventsEmit(m.ctx, "download_progress", Progress{
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
	units := []string{"B", "KB", "MB", "GB"}
	val := float64(b)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", val, units[i])
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
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := m.extractFile(dest, f)
		if err != nil {
			return err
		}
		percentage := (float64(i+1) / float64(totalFiles)) * 100
		wruntime.EventsEmit(m.ctx, "download_progress", Progress{
			Name:       name,
			Percentage: percentage,
			Status:     fmt.Sprintf("Extracting %d/%d...", i+1, totalFiles),
		})
	}

	wruntime.EventsEmit(m.ctx, "download_progress", Progress{
		Name:       name,
		Percentage: 100,
		Status:     "Completed",
	})
	return nil
}

func (m *Manager) extractFile(dest string, f *zip.File) error {
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
