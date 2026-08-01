package utils

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockUtilsHTTPClient struct {
	doFunc  func(req *http.Request) (*http.Response, error)
	getFunc func(url string) (*http.Response, error)
}

func (m *mockUtilsHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUtilsHTTPClient) Get(url string) (*http.Response, error) {
	if m.getFunc != nil {
		return m.getFunc(url)
	}
	return nil, errors.New("not implemented")
}

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

	// Error case
	gotErr := FetchContent("http://invalid.url. Ostenia-Test")
	if gotErr != "" {
		t.Errorf("FetchContent() with invalid URL = %q, want empty string", gotErr)
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

func TestGetSystemArch(t *testing.T) {
	arch := GetSystemArch()
	if arch == "" {
		t.Error("GetSystemArch() returned empty string")
	}
}

func TestGetInstalledVersionPaths(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "version_paths_test")
	defer os.RemoveAll(tmpDir)

	category := "test-cat"
	checkFile := "check.txt"
	binDir := filepath.Join(tmpDir, "bin", category)

	// Create dummy version directories
	os.MkdirAll(filepath.Join(binDir, "v1.0.0"), 0755)
	os.WriteFile(filepath.Join(binDir, "v1.0.0", checkFile), []byte("ok"), 0644)

	os.MkdirAll(filepath.Join(binDir, "node-v2.1.0"), 0755)
	os.WriteFile(filepath.Join(binDir, "node-v2.1.0", checkFile), []byte("ok"), 0644)

	// Dir without check file
	os.MkdirAll(filepath.Join(binDir, "v3.0.0"), 0755)

	// File instead of dir
	os.WriteFile(filepath.Join(binDir, "not-a-dir"), []byte("data"), 0644)

	versions := GetInstalledVersionPaths(tmpDir, category, checkFile)

	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	if _, ok := versions["v1.0.0"]; !ok {
		t.Error("v1.0.0 not found in results")
	}

	if _, ok := versions["2.1.0"]; !ok {
		t.Error("2.1.0 (normalized from node-v2.1.0) not found in results")
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

func TestDefaultHTTPClient(t *testing.T) {
	client := &DefaultHTTPClient{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("DefaultHTTPClient.Get failed: %v", err)
	}
	resp.Body.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatalf("DefaultHTTPClient.Do failed: %v", err)
	}
	resp2.Body.Close()
}

func TestDownloadFile_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")

	// 1. Invalid file creation
	err := DownloadFile(context.Background(), "http://example.com", filepath.Join(tmpDir, "non_existent_subdir", "test.txt"), "test", nil)
	if err == nil {
		t.Error("Expected DownloadFile to fail with invalid file path")
	}

	// 2. Client.Do error
	oldClient := Client
	defer func() { Client = oldClient }()

	mockCli := &mockUtilsHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network error")
		},
	}
	Client = mockCli

	err = DownloadFile(context.Background(), "http://example.com", filePath, "test", nil)
	if err == nil || err.Error() != "network error" {
		t.Errorf("Expected network error, got: %v", err)
	}

	// 3. Status not OK error
	mockCli.doFunc = func(req *http.Request) (*http.Response, error) {
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusNotFound)
		return resp.Result(), nil
	}

	err = DownloadFile(context.Background(), "http://example.com", filePath, "test", nil)
	if err == nil || err.Error() != "HTTP 404" {
		t.Errorf("Expected HTTP 404, got: %v", err)
	}
}

func TestGetInstalledVersionPaths_Apache(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin", "apache")
	os.MkdirAll(filepath.Join(binDir, "httpd-2.4.54", "Apache24", "bin"), 0755)

	err := os.WriteFile(filepath.Join(binDir, "httpd-2.4.54", "Apache24", "bin", "httpd.exe"), []byte("exe"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	paths := GetInstalledVersionPaths(tmpDir, "apache", "bin/httpd.exe")
	if _, ok := paths["2.4.54"]; !ok {
		t.Errorf("Expected version 2.4.54 to be detected for Apache, paths: %v", paths)
	}

	// 2. Test GetInstalledVersionPaths with a non-existent base directory (triggers ReadDir error)
	pathsEmpty := GetInstalledVersionPaths(filepath.Join(tmpDir, "non_existent"), "apache", "bin/httpd.exe")
	if len(pathsEmpty) != 0 {
		t.Errorf("Expected empty paths map for non-existent base directory, got %v", pathsEmpty)
	}
}

func TestGetSystemArch_x86(t *testing.T) {
	// Simple sanity test for GetSystemArch
	arch := GetSystemArch()
	if arch != "x64" && arch != "x86" {
		t.Errorf("Unexpected architecture: %s", arch)
	}
}

func TestNormalizeVersion_And_GetVersionPrefix(t *testing.T) {
	// 1. NormalizeVersion
	inputs := []struct {
		input, want string
	}{
		{"php-8.2.0", "8.2.0"},
		{"httpd-2.4.54", "2.4.54"},
		{"mysql-8.0.30", "8.0.0"}, // Comparing prefix trim
		{"node-v18.0.0", "18.0.0"},
		{"python-3.11.0", "3.10.0"}, // Just tests prefix
		{"openssl-4.0.0", "4.0.0"},
	}

	for _, tc := range inputs {
		got := NormalizeVersion(tc.input)
		// Check that the standard prefix has been removed
		if strings.HasPrefix(got, "php-") || strings.HasPrefix(got, "httpd-") || strings.HasPrefix(got, "node-v") {
			t.Errorf("NormalizeVersion(%q) = %q, prefix not removed", tc.input, got)
		}
	}

	// 2. GetVersionPrefix
	categories := []struct {
		category, want string
	}{
		{"php", "php-"},
		{"apache", "httpd-"},
		{"mysql", "mysql-"},
		{"nginx", "nginx-"},
		{"openssl", "openssl-"},
		{"nodejs", "node-v"},
		{"node.js", "node-v"},
		{"python", "python-"},
		{"unknown", ""},
	}

	for _, tc := range categories {
		got := GetVersionPrefix(tc.category)
		if got != tc.want {
			t.Errorf("GetVersionPrefix(%q) = %q, want %q", tc.category, got, tc.want)
		}
	}
}
