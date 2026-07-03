package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ModuleDefinition struct {
	Name      string
	CheckFile string
}

// DownloadFile downloads a file from URL to the specified path with context and progress report.
func DownloadFile(ctx context.Context, url, path, name string, onProgress func(percentage float64, status, speed, downloaded string)) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := Client.Do(req)
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
			if onProgress != nil {
				onProgress(pct, "Downloading...", s, FormatBytes(p))
			}
		},
	}
	_, err = io.Copy(out, io.TeeReader(resp.Body, wc))
	return err
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

func FormatBytes(b uint64) string {
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

// GetSystemArch returns the architecture string for the current system ("x64" or "x86").
func GetSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

// HTTPClient is an interface for making HTTP GET requests, allowing for test mocking.
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient is the production implementation of HTTPClient.
type DefaultHTTPClient struct{}

func (c *DefaultHTTPClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

// Client is the global HTTP client instance.
var Client HTTPClient = &DefaultHTTPClient{}

// FetchContent retrieves the content of a URL as a string.
func FetchContent(url string) string {
	resp, err := Client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// CompareVersions compares two version strings (major.minor.patch).
func CompareVersions(v1, v2 string) int {
	var a1, b1, c1 int
	var a2, b2, c2 int
	fmt.Sscanf(v1, "%d.%d.%d", &a1, &b1, &c1)
	fmt.Sscanf(v2, "%d.%d.%d", &a2, &b2, &c2)
	if a1 != a2 {
		return a1 - a2
	}
	if b1 != b2 {
		return b1 - b2
	}
	return c1 - c2
}

// GetInstalledVersionPaths returns a map of version strings to their absolute paths for a given category.
// It scans the bin/[category] directory for subfolders containing the specified check file.
func GetInstalledVersionPaths(baseDir, category, checkFile string) map[string]string {
	versions := make(map[string]string)
	binDir := filepath.Join(baseDir, "bin", category)
	entries, err := os.ReadDir(binDir)
	if err != nil {
		return versions
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			checkPath := filepath.Join(binDir, entry.Name(), checkFile)

			// Special handling for Apache's directory structure
			if category == "apache" {
				if _, err := os.Stat(checkPath); os.IsNotExist(err) {
					checkPath = filepath.Join(binDir, entry.Name(), "Apache24", checkFile)
				}
			}

			if _, err := os.Stat(checkPath); err == nil {
				// Normalize version string by removing common prefixes
				v := entry.Name()
				v = strings.TrimPrefix(v, "php-")
				v = strings.TrimPrefix(v, "httpd-")
				v = strings.TrimPrefix(v, "mysql-")
				v = strings.TrimPrefix(v, "nginx-")
				v = strings.TrimPrefix(v, "node-v")
				v = strings.TrimPrefix(v, "python-")
				v = strings.TrimPrefix(v, "heidisql-")
				versions[v] = filepath.Join(binDir, entry.Name())
			}
		}
	}
	return versions
}
