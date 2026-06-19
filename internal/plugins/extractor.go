package plugins

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip extracts a zip file to a destination directory.
func Unzip(ctx context.Context, src string, dest string, name string, emit func(context.Context, string, ...interface{})) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	totalFiles := len(r.File)
	for i, f := range r.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := extractFile(dest, f); err != nil {
				return err
			}

			// Report extraction progress
			if i%20 == 0 || i+1 == totalFiles {
				percentage := (float64(i+1) / float64(totalFiles)) * 100
				emit(ctx, "download_progress", Progress{
					Name:       name,
					Percentage: percentage,
					Status:     "Extracting...",
				})
			}
		}
	}
	return nil
}

func extractFile(dest string, f *zip.File) error {
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

	return writeFile(fpath, f)
}

func writeFile(fpath string, f *zip.File) error {
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
