package plugins

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unzip_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "dest")
	os.MkdirAll(destDir, 0755)

	// Create a dummy zip file
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	zw := zip.NewWriter(f)

	cf, err := zw.Create("hello.txt")
	if err != nil {
		t.Fatalf("failed to create file in zip: %v", err)
	}
	cf.Write([]byte("hello world"))

	zw.Close()
	f.Close()

	err = Unzip(context.Background(), zipPath, destDir, "test", func(ctx context.Context, s string, i ...interface{}) {})
	if err != nil {
		t.Fatalf("Unzip failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "hello.txt"))
	if err != nil {
		t.Fatalf("failed to read unzipped file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("Expected hello world, got %s", string(content))
	}
}
