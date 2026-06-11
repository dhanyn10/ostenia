package plugins

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadAndExtract(t *testing.T) {
	// 1. Create a dummy ZIP file in memory
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// Add a file to the ZIP
	f, err := zw.Create("test_plugin/bin/test.exe")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("dummy executable content"))
	if err != nil {
		t.Fatal(err)
	}
	zw.Close()

	// 2. Set up a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(buf.Bytes())
	}))
	defer ts.Close()

	// 3. Prepare environment
	tmpBaseDir, _ := os.MkdirTemp("", "ostenia_base_test")
	defer os.RemoveAll(tmpBaseDir)

	mgr := NewManager(context.Background())

	t.Run("Unzip Logic", func(t *testing.T) {
		zipFile := filepath.Join(tmpBaseDir, "test.zip")
		os.WriteFile(zipFile, buf.Bytes(), 0644)

		extractDir := filepath.Join(tmpBaseDir, "extracted")
		err := mgr.unzipFile(context.Background(), zipFile, extractDir, "Test")
		if err != nil {
			t.Fatalf("unzipFile failed: %v", err)
		}

		expectedFile := filepath.Join(extractDir, "test_plugin/bin/test.exe")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not extracted", expectedFile)
		}
	})

	t.Run("Mock Download Logic", func(t *testing.T) {
		tmpFile := filepath.Join(tmpBaseDir, "downloaded.zip")
		err := mgr.downloadFile(context.Background(), ts.URL, tmpFile, "TestDownload")
		if err != nil {
			t.Fatalf("downloadFile failed: %v", err)
		}

		if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
			t.Error("Downloaded file does not exist")
		}

		// Verify content length
		fi, _ := os.Stat(tmpFile)
		if fi.Size() == 0 {
			t.Error("Downloaded file is empty")
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.bytes); got != tt.want {
			t.Errorf("formatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
		}
	}
}
