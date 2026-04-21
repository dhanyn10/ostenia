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
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func getSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

func detectAllPHPVersions() ([]string, string) {
	arch := getSystemArch()
	baseURL := "https://downloads.php.net/~windows/releases/"

	resp, err := http.Get(baseURL)
	if err != nil {
		return []string{"8.3.6"}, baseURL
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	re := regexp.MustCompile(`php-(8\.\d+\.\d+)-Win32-vs16-` + arch + `\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			seen[v] = true
		}
	}

	if len(versions) == 0 {
		return []string{"8.3.6"}, baseURL
	}
	
	// Reverse versions to get newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}
	
	return versions, baseURL
}

func detectAllApacheVersions() ([]string, map[string]string) {
	arch := getSystemArch()
	baseURL := "https://www.apachelounge.com/download/"
	
	// Updated base for binaries
	binaryBaseURL := "https://www.apachelounge.com/download/VS18/binaries/"
	
	// Default fallbacks with correct new paths
	defaultVer := "2.4.66-260223"
	defaultURL := binaryBaseURL + "httpd-2.4.66-260223-Win64-VS18.zip"
	if arch == "x86" {
		defaultVer = "2.4.66-260131"
		defaultURL = "https://www.apachelounge.com/download/vs18/binaries/httpd-2.4.66-260131-win32-vs18.zip"
	}

	resp, err := http.Get(baseURL)
	if err != nil {
		return []string{defaultVer}, map[string]string{defaultVer: defaultURL}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Regex to find httpd-2.4.x-y-Win64-VSz.zip or win32
	var re *regexp.Regexp
	if arch == "x64" {
		re = regexp.MustCompile(`httpd-(2\.4\.\d+-\d+)-Win64-VS\d+\.zip`)
	} else {
		re = regexp.MustCompile(`httpd-(2\.4\.\d+-\d+)-win32-vs\d+\.zip`)
	}
	
	matches := re.FindAllStringSubmatch(content, -1)
	
	versions := []string{}
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			// Constructing the correct URL based on the pattern
			if arch == "x64" {
				urlMap[v] = binaryBaseURL + m[0]
			} else {
				urlMap[v] = "https://www.apachelounge.com/download/vs18/binaries/" + m[0]
			}
			seen[v] = true
		}
	}

	if len(versions) == 0 {
		return []string{defaultVer}, map[string]string{defaultVer: defaultURL}
	}

	return versions, urlMap
}

type DownloadTask struct {
	Name          string            `json:"name"`
	URL           string            `json:"url"`
	Version       string            `json:"version"`
	Versions      []string          `json:"versions"`      // Optional: available versions
	VersionURLs   map[string]string `json:"versionUrls"`   // Map version to URL
	InstalledVers []string          `json:"installedVers"` // Local installed versions
	Target        string            `json:"target"`        // relative to bin/
	CheckFile     string            `json:"checkFile"`     // file that must exist to verify installation
	IsInstalled   bool              `json:"isInstalled"`
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

func detectAllMySQLVersions() ([]string, map[string]string) {
	arch := getSystemArch()
	mysqlFileArch := "winx64"
	if arch == "x86" {
		mysqlFileArch = "win32"
	}

	baseURL := "https://downloads.mysql.com/archives/community/"
	
	// Default fallbacks
	defaultVers := []string{"8.0.37", "8.4.0", "9.1.0"}
	urlMap := make(map[string]string)
	for _, v := range defaultVers {
		urlMap[v] = fmt.Sprintf("https://downloads.mysql.com/archives/get/p/23/file/mysql-%s-%s.zip", v, mysqlFileArch)
	}

	resp, err := http.Get(baseURL)
	if err != nil {
		return defaultVers, urlMap
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	// Regex for mysql-8.0.37-winx64.zip
	re := regexp.MustCompile(`mysql-(\d+\.\d+\.\d+)-` + mysqlFileArch + `\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = fmt.Sprintf("https://downloads.mysql.com/archives/get/p/23/file/mysql-%s-%s.zip", v, mysqlFileArch)
			seen[v] = true
		}
	}

	if len(versions) == 0 {
		return defaultVers, urlMap
	}

	// Reverse to get newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	return versions, urlMap
}

// GetLatestKnownVersions returns a list of prerequisite downloads
func GetLatestKnownVersions() []DownloadTask {
	phpVers, baseURL := detectAllPHPVersions()
	apacheVers, apacheURLs := detectAllApacheVersions()
	mysqlVers, mysqlURLs := detectAllMySQLVersions()
	arch := getSystemArch()

	latestPHP := phpVers[0] // Now newest is first
	phpURL := baseURL + fmt.Sprintf("php-%s-Win32-vs16-%s.zip", latestPHP, arch)

	latestApache := apacheVers[0]
	apacheURL := apacheURLs[latestApache]

	latestMySQL := mysqlVers[0]
	mysqlURL := mysqlURLs[latestMySQL]

	tasks := []DownloadTask{
		{
			Name:     "PHP",
			URL:      phpURL,
			Version:  latestPHP,
			Versions: phpVers,
			Target:   "php/php-" + latestPHP,
			CheckFile: "php.exe",
		},
		{
			Name:        "Apache",
			URL:         apacheURL,
			Version:     latestApache,
			Versions:    apacheVers,
			VersionURLs: apacheURLs,
			Target:      "apache/httpd-" + latestApache,
			CheckFile:   "bin/httpd.exe",
		},
		{
			Name:        "MySQL",
			URL:         mysqlURL,
			Version:     latestMySQL,
			Versions:    mysqlVers,
			VersionURLs: mysqlURLs,
			Target:      "mysql/mysql-" + latestMySQL,
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


	for i := range tasks {
		baseDir := config.GetBaseDir()
		tasks[i].InstalledVers = []string{}
		
		targetParts := strings.Split(tasks[i].Target, "/")
		if len(targetParts) == 2 {
			compDir := filepath.Join(baseDir, "bin", targetParts[0])
			entries, err := os.ReadDir(compDir)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() && entry.Name() != "current" {
						// Extract version from folder name
						// If name is "httpd-2.4.66", ver is "2.4.66"
						// If name is "php-8.2.0", ver is "8.2.0"
						ver := entry.Name()
						dashIdx := strings.Index(ver, "-")
						if dashIdx != -1 {
							ver = ver[dashIdx+1:]
						}

						// Verify check file (handle potential wrapping like Apache24/)
						checkPaths := []string{
							filepath.Join(compDir, entry.Name(), tasks[i].CheckFile),
						}
						
						// Add common wrapped paths
						if tasks[i].Name == "Apache" {
							checkPaths = append(checkPaths, filepath.Join(compDir, entry.Name(), "Apache24", tasks[i].CheckFile))
						} else if tasks[i].Name == "MySQL" {
							// MySQL often wraps in a folder named after the zip
							// but since we don't know the exact name easily here, 
							// we'll rely on the unwrap logic during extraction.
							// However, let's try to detect if bin/mysqld.exe is one level deeper
							subEntries, _ := os.ReadDir(filepath.Join(compDir, entry.Name()))
							for _, sub := range subEntries {
								if sub.IsDir() {
									checkPaths = append(checkPaths, filepath.Join(compDir, entry.Name(), sub.Name(), tasks[i].CheckFile))
								}
							}
						}

						found := false
						for _, cp := range checkPaths {
							if _, err := os.Stat(cp); err == nil {
								found = true
								break
							}
						}

						if found {
							tasks[i].InstalledVers = append(tasks[i].InstalledVers, ver)
						}
					}
				}
			}
		} else {
			// HeidiSQL case
			checkPath := filepath.Join(baseDir, "bin", tasks[i].Target, tasks[i].CheckFile)
			if _, err := os.Stat(checkPath); err == nil {
				tasks[i].InstalledVers = append(tasks[i].InstalledVers, tasks[i].Version)
			}
		}

		// Verification for the currently targeted version folder
		// We should also check for wrapped folders here
		checkPath := filepath.Join(baseDir, "bin", tasks[i].Target, tasks[i].CheckFile)
		if _, err := os.Stat(checkPath); err == nil {
			tasks[i].IsInstalled = true
		} else if tasks[i].Name == "Apache" {
			if _, err := os.Stat(filepath.Join(baseDir, "bin", tasks[i].Target, "Apache24", tasks[i].CheckFile)); err == nil {
				tasks[i].IsInstalled = true
			}
		}
	}

	return tasks
}

func (m *Manager) DeleteVersion(taskName, version string) error {
	baseDir := config.GetBaseDir()
	prefix := ""
	switch strings.ToLower(taskName) {
	case "php":
		prefix = "php/php"
	case "apache":
		prefix = "apache/httpd"
	case "mysql":
		prefix = "mysql/mysql"
	case "heidisql":
		prefix = "heidisql"
	default:
		return fmt.Errorf("unknown task name")
	}

	// Try both prefix-version and just version if prefix is empty or if we can't find it
	targetDir := filepath.Join(baseDir, "bin", prefix+"-"+version)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// Try without prefix part of the folder name if it was detected differently
		category := strings.Split(prefix, "/")[0]
		targetDir = filepath.Join(baseDir, "bin", category, version)
	}

	return os.RemoveAll(targetDir)
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

	// Robust installed check (including wrapped)
	isAlreadyInstalled := false
	if _, err := os.Stat(filepath.Join(targetDir, task.CheckFile)); err == nil {
		isAlreadyInstalled = true
	} else if task.Name == "Apache" {
		if _, err := os.Stat(filepath.Join(targetDir, "Apache24", task.CheckFile)); err == nil {
			isAlreadyInstalled = true
		}
	}

	if isAlreadyInstalled {
		fmt.Printf("[Downloader] %s verified at %s\n", task.Name, targetDir)
		m.ensureCurrentLink(task)
		wruntime.EventsEmit(m.ctx, "download_progress", Progress{
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
			wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Cancelled"})
		} else {
			wruntime.EventsEmit(m.ctx, "download_error", map[string]string{"name": task.Name, "error": "Download failed: " + err.Error()})
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
			wruntime.EventsEmit(m.ctx, "download_progress", Progress{Name: task.Name, Percentage: 0, Status: "Cancelled"})
		} else {
			wruntime.EventsEmit(m.ctx, "download_error", map[string]string{"name": task.Name, "error": "Extraction failed: " + err.Error()})
		}
		return err
	}

	// AUTO-UNWRAP: If there is only one directory in extractTmp, move its contents up
	entries, _ := os.ReadDir(extractTmp)
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(extractTmp, entries[0].Name())
		fmt.Printf("[Downloader] Unwrapping %s\n", subDir)
		subEntries, _ := os.ReadDir(subDir)
		for _, se := range subEntries {
			os.Rename(filepath.Join(subDir, se.Name()), filepath.Join(extractTmp, se.Name()))
		}
		os.Remove(subDir)
	}

	// Rename to final targetDir
	os.RemoveAll(targetDir)
	err = os.Rename(extractTmp, targetDir)
	if err != nil {
		fmt.Printf("[Downloader] Rename error: %v, trying manual move if cross-device\n", err)
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
