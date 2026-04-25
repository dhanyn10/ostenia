package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Progress struct for download progress updates
type Progress struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Status     string  `json:"status"`
	Speed      string  `json:"speed"`
	Downloaded string  `json:"downloaded"`
}

// WriteCounter counts the number of bytes written and reports progress.
type WriteCounter struct {
	Total      uint64
	Downloaded uint64
	StartTime  time.Time
	OnProgress func(progress uint64, total uint64, speed string)
}

// Write implements the io.Writer interface.
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

// formatBytes converts bytes to a human-readable format (e.g., KB, MB).
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

// DownloadFileWithContext downloads a file from a URL with context for cancellation and progress reporting.
func DownloadFileWithContext(ctx context.Context, url string, filepath string, name string) error {
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
			wruntime.EventsEmit(ctx, "download_progress", Progress{
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
