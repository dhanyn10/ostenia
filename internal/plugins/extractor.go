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
			fpath := filepath.Join(dest, f.Name)
			if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
				return fmt.Errorf("%s: illegal file path", fpath)
			}

			if f.FileInfo().IsDir() {
				os.MkdirAll(fpath, os.ModePerm)
				continue
			}

			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			rc, err := f.Open()
			if err != nil {
				outFile.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
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
