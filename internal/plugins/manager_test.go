package plugins

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"ostenia/internal/plugins/utils"
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
	// Mock the emitter to avoid Wails runtime errors during tests
	mgr.emit = func(ctx context.Context, eventName string, optionalData ...interface{}) {}

	t.Run("Unzip Logic", func(t *testing.T) {
		zipFile := filepath.Join(tmpBaseDir, "test.zip")
		os.WriteFile(zipFile, buf.Bytes(), 0644)

		extractDir := filepath.Join(tmpBaseDir, "extracted")
		err := Unzip(context.Background(), zipFile, extractDir, "Test", mgr.emit)
		if err != nil {
			t.Fatalf("Unzip failed: %v", err)
		}

		expectedFile := filepath.Join(extractDir, "test_plugin/bin/test.exe")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not extracted", expectedFile)
		}
	})

	t.Run("Mock Download Logic", func(t *testing.T) {
		tmpFile := filepath.Join(tmpBaseDir, "downloaded.zip")
		err := utils.DownloadFile(context.Background(), ts.URL, tmpFile, "TestDownload", nil)
		if err != nil {
			t.Fatalf("DownloadFile failed: %v", err)
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
		if got := utils.FormatBytes(tt.bytes); got != tt.want {
			t.Errorf("FormatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
		}
	}
}
