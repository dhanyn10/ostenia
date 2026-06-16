package plugins

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func Unzip(ctx context.Context, src string, dest string, name string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	totalFiles := len(r.File)
	for i, f := range r.File {
		if err := extractZipFile(ctx, f, dest, name, i, totalFiles); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(ctx context.Context, f *zip.File, dest, name string, index, total int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			return os.MkdirAll(fpath, os.ModePerm)
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		if err := copyZipContent(f, fpath); err != nil {
			return err
		}

		percentage := (float64(index+1) / float64(total)) * 100
		wruntime.EventsEmit(ctx, "download_progress", Progress{
			Name:       name,
			Percentage: percentage,
			Status:     "Extracting...",
		})
		return nil
	}
}

func copyZipContent(f *zip.File, fpath string) error {
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(outFile, rc)
	return err
}
