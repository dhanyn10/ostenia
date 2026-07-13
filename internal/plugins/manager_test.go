package plugins

import (
	"bytes"
	"context"
	"fmt"
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
	err     error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
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
		oldEnv := os.Getenv("OSTENIA_HOME")
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Setenv("OSTENIA_HOME", oldEnv)

		verDir := filepath.Join(tempDir, "bin", "php", "php-8.1.0")
		os.MkdirAll(verDir, 0755)

		_ = m.DeleteVersion("PHP", "8.1.0")

		// Test with Node.js name mapping
		nodeDir := filepath.Join(tempDir, "bin", "nodejs", "node-v18.0.0")
		os.MkdirAll(nodeDir, 0755)
		_ = m.DeleteVersion("Node.js", "18.0.0")

		// Test Windows branch
		origGOOS := RuntimeGOOS
		defer func() { RuntimeGOOS = origGOOS }()
		RuntimeGOOS = "windows"
		_ = m.DeleteVersion("PHP", "8.2.0")
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
		oldEnv := os.Getenv("OSTENIA_HOME")
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Setenv("OSTENIA_HOME", oldEnv)

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
		DetectHeidiSQLInstallationOverride = func() (string, string) {
			return "/usr/bin/heidisql.exe", ""
		}
		defer func() { DetectHeidiSQLInstallationOverride = nil }()

		// Create modules folders
		os.MkdirAll(filepath.Join(phpDir, "composer"), 0755)
		os.WriteFile(filepath.Join(phpDir, "composer.phar"), []byte(""), 0644)

        _ = GetLatestKnownVersions()
    })

	t.Run("DownloadAndExtract_Basic", func(t *testing.T) {
		tempDir := t.TempDir()
		oldEnv := os.Getenv("OSTENIA_HOME")
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Setenv("OSTENIA_HOME", oldEnv)

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

		// Test download error path
		utils.Client = &mockHTTPClient{err: fmt.Errorf("download error")}
		err = m.DownloadAndExtract(task)
		if err == nil {
			t.Error("Expected download error")
		}
	})

	t.Run("GetInstalledVersionPaths_Extended", func(t *testing.T) {
		tempDir := t.TempDir()
		_ = m.GetInstalledVersionPaths("php", "php.exe")
		_ = tempDir
	})

	t.Run("ModuleMethods", func(t *testing.T) {
		tempDir := t.TempDir()
		phpPath := filepath.Join(tempDir, "bin", "php", "current")
		os.MkdirAll(phpPath, 0755)
		os.WriteFile(filepath.Join(phpPath, "php.exe"), []byte(""), 0755)

		_ = m.InstallModule("Composer", phpPath, nil)
		_ = m.UninstallModule("Composer", phpPath)

		_ = m.InstallModule("Xdebug", phpPath, nil)
		_ = m.UninstallModule("Xdebug", phpPath)
	})

	t.Run("DownloadFileManual", func(t *testing.T) {
		utils.Client = &mockHTTPClient{content: "manual"}
		tempDir := t.TempDir()
		dest := filepath.Join(tempDir, "manual.exe")
		err := m.DownloadFileManual("http://example.com/manual.exe", dest, "Manual")
		if err != nil {
			t.Errorf("DownloadFileManual failed: %v", err)
		}
	})

	t.Run("handleArchive_Mocked", func(t *testing.T) {
		tempDir := t.TempDir()
		oldEnv := os.Getenv("OSTENIA_HOME")
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Setenv("OSTENIA_HOME", oldEnv)

		oldUnzip := unzipFunc
		unzipFunc = func(ctx context.Context, src, dest, name string, emit func(context.Context, string, ...interface{})) error {
			os.MkdirAll(filepath.Join(dest, "sub"), 0755)
			os.WriteFile(filepath.Join(dest, "sub", "test.txt"), []byte("data"), 0644)
			return nil
		}
		defer func() { unzipFunc = oldUnzip }()

		task := DownloadTask{
			Name: "TestZip",
			Version: "1.0.0",
			URL: "http://example.com/test.zip",
			Target: "testzip/1.0.0",
			CheckFile: "sub/test.txt",
		}
		targetDir := filepath.Join(tempDir, "bin", task.Target)

		err := m.handleArchive(task, "dummy.zip", targetDir)
		if err != nil {
			t.Errorf("handleArchive failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(targetDir, "test.txt")); os.IsNotExist(err) {
			t.Error("Expected test.txt to be moved to targetDir from sub folder")
		}

		// Test Nupkg path
		task.URL = "http://example.com/test.nupkg"
		unzipFunc = func(ctx context.Context, src, dest, name string, emit func(context.Context, string, ...interface{})) error {
			os.MkdirAll(filepath.Join(dest, "tools"), 0755)
			os.WriteFile(filepath.Join(dest, "tools", "app.exe"), []byte("data"), 0644)
			return nil
		}
		err = m.handleArchive(task, "dummy.nupkg", targetDir+"_nupkg")
		if err != nil {
			t.Errorf("handleArchive nupkg failed: %v", err)
		}
	})

	t.Run("NewManager_Emit", func(t *testing.T) {
		mgr := NewManager(context.Background())
		// Directly test the emit logic without calling wruntime
		mgr.emit = func(ctx context.Context, eventName string, optionalData ...interface{}) {}
		mgr.emit(context.Background(), "test-event")
		mgr.emit(nil, "test-event")
	})

	t.Run("DownloadAndExtract_AlreadyInstalled", func(t *testing.T) {
		tempDir := t.TempDir()
		oldEnv := os.Getenv("OSTENIA_HOME")
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Setenv("OSTENIA_HOME", oldEnv)

		task := DownloadTask{
			Name: "PHP",
			Version: "8.2.0",
			Target: "php/php-8.2.0",
			CheckFile: "php.exe",
		}
		targetDir := filepath.Join(tempDir, "bin", task.Target)
		os.MkdirAll(targetDir, 0755)
		os.WriteFile(filepath.Join(targetDir, "php.exe"), []byte(""), 0644)

		err := m.DownloadAndExtract(task)
		if err != nil {
			t.Errorf("DownloadAndExtract should have skipped and returned nil, got %v", err)
		}

		// Test already installed branch with Apache special case
		task2 := DownloadTask{
			Name: "Apache",
			Version: "2.4.54",
			Target: "apache/httpd-2.4.54",
			CheckFile: "bin/httpd.exe",
		}
		targetDir2 := filepath.Join(tempDir, "bin", task2.Target)
		os.MkdirAll(filepath.Join(targetDir2, "Apache24", "bin"), 0755)
		os.WriteFile(filepath.Join(targetDir2, "Apache24", "bin", "httpd.exe"), []byte(""), 0644)
		if !m.isAlreadyInstalled(task2, targetDir2) {
			t.Error("isAlreadyInstalled failed for Apache special case")
		}
	})
}
