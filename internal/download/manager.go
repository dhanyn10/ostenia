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

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DownloadTask struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
	Target  string `json:"target"` // relative to bin/
}

type Progress struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
}

type Manager struct {
	ctx context.Context
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{ctx: ctx}
}

// GetLatestKnownVersions returns a list of prerequisite downloads
func GetLatestKnownVersions() []DownloadTask {
	return []DownloadTask{
		{
			Name:    "PHP",
			URL:     "https://windows.php.net/downloads/releases/php-8.3.4-Win32-vs16-x64.zip",
			Version: "8.3.4",
			Target:  "php/php-8.3.4",
		},
		{
			Name:    "Apache",
			URL:     "https://www.apachelounge.com/download/VS17/binaries/httpd-2.4.58-win64-VS17.zip",
			Version: "2.4.58",
			Target:  "apache/httpd-2.4.58",
		},
		{
			Name:    "MySQL",
			URL:     "https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.36-winx64.zip",
			Version: "8.0.36",
			Target:  "mysql/mysql-8.0.36",
		},
		{
			Name:    "HeidiSQL",
			URL:     "https://www.heidisql.com/downloads/releases/HeidiSQL_12.6_64_Portable.zip",
			Version: "12.6",
			Target:  "heidisql",
		},
	}
}

func (m *Manager) DownloadAndExtract(task DownloadTask) error {
	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin")
	targetDir := filepath.Join(binDir, task.Target)

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		runtime.EventsEmit(m.ctx, "download_progress", Progress{
			Name:       task.Name,
			Percentage: 100,
			Status:     "Already installed",
		})
		return nil
	}

	err := os.MkdirAll(binDir, 0755)
	if err != nil {
		return err
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.zip", task.Name))
	err = m.downloadFile(task.URL, tmpFile, task.Name)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	return m.unzip(tmpFile, targetDir, task.Name)
}

func (m *Manager) downloadFile(url string, filepath string, name string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	counter := &WriteCounter{
		Total: uint64(resp.ContentLength),
		OnProgress: func(progress uint64, total uint64) {
			percentage := float64(progress) / float64(total) * 100
			runtime.EventsEmit(m.ctx, "download_progress", Progress{
				Name:       name,
				Percentage: percentage,
				Status:     "Downloading...",
			})
		},
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	return err
}

func (m *Manager) unzip(src string, dest string, name string) error {
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
		err := m.extractAndWriteFile(dest, f)
		if err != nil {
			return err
		}
		percentage := float64(i+1) / float64(totalFiles) * 100
		runtime.EventsEmit(m.ctx, "download_progress", Progress{
			Name:       name,
			Percentage: percentage,
			Status:     "Extracting...",
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
	OnProgress func(progress uint64, total uint64)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += uint64(n)
	wc.OnProgress(wc.Downloaded, wc.Total)
	return n, nil
}
