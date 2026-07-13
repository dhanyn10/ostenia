package plugins

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUnzip(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")
	destDir := filepath.Join(tempDir, "dest")

	// Create a dummy zip file
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	f, err := zw.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}

	// Add a directory
	_, err = zw.Create("subdir/")
	if err != nil {
		t.Fatal(err)
	}

	err = zw.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test Unzip
	emit := func(ctx context.Context, eventName string, optionalData ...interface{}) {}
	err = Unzip(context.Background(), zipPath, destDir, "test", emit)
	if err != nil {
		t.Errorf("Unzip failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
	if err != nil {
		t.Errorf("Failed to read extracted file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(content))
	}
}

func TestUnzip_Security(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "malicious.zip")
	destDir := filepath.Join(tempDir, "dest")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// Malicious path: directory traversal
	f, _ := zw.Create("../outside.txt")
	f.Write([]byte("malicious"))
	zw.Close()
	os.WriteFile(zipPath, buf.Bytes(), 0644)

	emit := func(ctx context.Context, eventName string, optionalData ...interface{}) {}
	err := Unzip(context.Background(), zipPath, destDir, "test", emit)
	if err == nil {
		t.Error("Expected error for directory traversal")
	}
}
