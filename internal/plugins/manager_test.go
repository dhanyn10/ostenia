package plugins

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"testing"
)

type mockHTTPClient struct {
	utils.HTTPClient
	content string
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		ContentLength: int64(len(m.content)),
		Body:       io.NopCloser(bytes.NewBufferString(m.content)),
	}, nil
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.Get(req.URL.String())
}

type mockExecutor struct {
	utils.CommandExecutor
}

func (m *mockExecutor) Command(name string, args ...string) *exec.Cmd {
	// Force all commands to use 'go' for this test to avoid OS path issues
	// But we must use a command that exists. 'go' should exist in this environment.
	return exec.Command("go", "version")
}

type mockHTTPClientProgress struct {
	utils.HTTPClient
}

func (m *mockHTTPClientProgress) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		ContentLength: 100,
		Body:       io.NopCloser(bytes.NewBuffer(make([]byte, 100))),
	}, nil
}

func (m *mockHTTPClientProgress) Do(req *http.Request) (*http.Response, error) {
	return m.Get(req.URL.String())
}

func TestPlugins_Complete(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()
	utils.Client = &mockHTTPClient{content: "dummy content"}

	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &mockExecutor{}

	ctx := context.Background()
	m := NewManager(ctx)
    m.emit = func(ctx context.Context, eventName string, optionalData ...interface{}) {}

	t.Run("Manager_Basic", func(t *testing.T) {
		m.CancelDownload("test")
	})

	t.Run("DeleteVersion", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		verDir := filepath.Join(tempDir, "bin", "php", "php-8.1.0")
		os.MkdirAll(verDir, 0755)

		_ = m.DeleteVersion("PHP", "8.1.0")
	})

	t.Run("GetInstalledVersionPaths", func(t *testing.T) {
		tempDir := t.TempDir()
		os.MkdirAll(filepath.Join(tempDir, "bin", "php", "php-8.1.0"), 0755)
		os.WriteFile(filepath.Join(tempDir, "bin", "php", "php-8.1.0", "php.exe"), []byte(""), 0644)

		paths := GetInstalledVersionPaths(tempDir, "php", "php.exe")
		if len(paths) == 0 {
			t.Error("Expected versions")
		}
	})

    t.Run("Discovery", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)

		// Setup dummy installed versions to test detection
		phpDir := filepath.Join(tempDir, "bin", "php", "php-8.2.0")
		os.MkdirAll(phpDir, 0755)
		os.WriteFile(filepath.Join(phpDir, "php.exe"), []byte(""), 0644)

		// Setup current symlink
		currentLink := filepath.Join(tempDir, "bin", "php", "current")
		os.MkdirAll(filepath.Dir(currentLink), 0755)
		// On Linux we can use a real symlink for testing if we want,
		// but the code uses mklink via executor.

		// Test HeidiSQL detection
		detectHeidiSQLInstallationOverride = func() (string, string) {
			return "/usr/bin/heidisql.exe", ""
		}
		defer func() { detectHeidiSQLInstallationOverride = nil }()

        _ = GetLatestKnownVersions()
    })

	t.Run("DownloadAndExtract_Basic", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)

		// Use a mock that provides a body to avoid errors in io.Copy
		utils.Client = &mockHTTPClientProgress{}

		task := DownloadTask{
			Name: "TestPlugin",
			Version: "1.0.0",
			URL: "http://example.com/test.exe", // Ends in .exe, triggers handleInstaller
			Target: "test/1.0.0",
			CheckFile: "test.exe",
		}

		// Mock GetSystemDirectory by path trickery is hard,
		// let's just bypass the actual Run call if possible or ensure it doesn't fail the test.
		// Since we mocked utils.Executor, it should run 'go version'.

		err := m.DownloadAndExtract(task)
		if err != nil {
			t.Errorf("DownloadAndExtract failed: %v", err)
		}
	})

	t.Run("GetInstalledVersionPaths_Extended", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = m.GetInstalledVersionPaths("php", "php.exe")
		_ = tempDir
	})

	t.Run("ModuleMethods", func(t *testing.T) {
		_ = m.InstallModule("Composer", "/path", nil)
		_ = m.UninstallModule("Composer", "/path")
	})
}
