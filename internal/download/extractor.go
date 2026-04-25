package download

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Unzip extracts a zip archive to a destination directory.
func Unzip(ctx context.Context, src string, dest string, name string) error {
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

		err := extractAndWriteFile(dest, f)
		if err != nil {
			return err
		}
		percentage := (float64(i+1) / float64(totalFiles)) * 100
		wruntime.EventsEmit(ctx, "download_progress", Progress{
			Name:       name,
			Percentage: percentage,
			Status:     fmt.Sprintf("Extracting %d/%d...", i+1, totalFiles),
		})
	}

	wruntime.EventsEmit(ctx, "download_progress", Progress{
		Name:       name,
		Percentage: 100,
		Status:     "Completed",
	})
	return nil
}

// extractAndWriteFile extracts a single file from a zip archive.
func extractAndWriteFile(dest string, f *zip.File) error {
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

// HandleExeDownload moves an exe and optionally runs it.
func HandleExeDownload(tmpFilePath string, targetDir string, checkFile string, runAfterMove bool) error {
	os.MkdirAll(targetDir, 0755)
	destFile := filepath.Join(targetDir, "installer.exe")

	if runtime.GOOS == "windows" {
		// Using 'cmd /c move' handles cross-drive moves automatically by doing copy+delete
		cmd := exec.Command("cmd", "/c", "move", "/Y", tmpFilePath, destFile)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("system move failed: %w", err)
		}
		if runAfterMove {
			exec.Command("cmd", "/c", "start", "", destFile).Run()
		}
	} else {
		// For non-windows, os.Rename is usually fine
		if err := os.Rename(tmpFilePath, destFile); err != nil {
			return fmt.Errorf("rename failed: %w", err)
		}
		if runAfterMove {
			// Attempt to run on non-windows, might need chmod +x first
			exec.Command(destFile).Start()
		}
	}
	return nil
}
