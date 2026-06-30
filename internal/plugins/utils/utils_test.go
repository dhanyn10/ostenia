package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.4", "1.2.3", 1},
		{"1.2.2", "1.2.3", -1},
		{"2.0.0", "1.9.9", 1},
		{"1.10.0", "1.2.0", 8},
		{"1.2.0", "1.10.0", -8},
	}

	for _, tt := range tests {
		if got := CompareVersions(tt.v1, tt.v2); got != tt.want {
			t.Errorf("CompareVersions(%q, %q) = %v, want %v", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestGetSystemDirectory(t *testing.T) {
	dir := GetSystemDirectory()
	if dir == "" {
		t.Error("GetSystemDirectory() returned empty string")
	}
}

func TestSafeEnv(t *testing.T) {
	env := SafeEnv()
	if len(env) == 0 {
		t.Error("SafeEnv() returned empty environment")
	}
	foundPath := false
	for _, e := range env {
		if len(e) >= 5 && e[:5] == "PATH=" {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Error("SafeEnv() did not contain PATH")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "copyfile_test")
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	want := "hello world"
	_ = os.WriteFile(src, []byte(want), 0644)

	err := CopyFile(src, dst)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	content, _ := os.ReadFile(dst)
	if string(content) != want {
		t.Errorf("Copied content = %q, want %q", string(content), want)
	}

	// Test non-existent source
	err = CopyFile("non-existent", dst)
	if err == nil {
		t.Error("CopyFile should fail for non-existent source")
	}
}

func TestFetchContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock content"))
	}))
	defer ts.Close()

	got := FetchContent(ts.URL)
	want := "mock content"
	if got != want {
		t.Errorf("FetchContent() = %q, want %q", got, want)
	}
}

func TestDownloadFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mock data"))
	}))
	defer ts.Close()

	tmpDir, _ := os.MkdirTemp("", "download_test")
	defer os.RemoveAll(tmpDir)
	tmpFile := filepath.Join(tmpDir, "test.txt")

	var progressCalled bool
	err := DownloadFile(context.Background(), ts.URL, tmpFile, "test", func(pct float64, status, speed, downloaded string) {
		progressCalled = true
	})

	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	if !progressCalled {
		t.Error("onProgress callback was not called")
	}

	content, _ := os.ReadFile(tmpFile)
	if string(content) != "mock data" {
		t.Errorf("Downloaded content = %q, want %q", string(content), "mock data")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		if got := FormatBytes(tt.bytes); got != tt.want {
			t.Errorf("FormatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
		}
	}
}
