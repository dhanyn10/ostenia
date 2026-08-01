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
		StatusCode:    http.StatusOK,
		ContentLength: int64(len(m.content)),
		Body:          io.NopCloser(bytes.NewBufferString(m.content)),
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
		StatusCode:    http.StatusOK,
		ContentLength: 100,
		Body:          io.NopCloser(bytes.NewBuffer(make([]byte, 100))),
	}, nil
}

func (m *mockHTTPClientProgress) Do(req *http.Request) (*http.Response, error) {
	return m.Get(req.URL.String())
}

func setupTestManager(t *testing.T) (*Manager, func()) {
	origClient := utils.Client
	origExecutor := utils.Executor
	utils.Client = &mockHTTPClient{content: "dummy content"}
	utils.Executor = &mockExecutor{}

	m := NewManager(context.Background())
	m.emit = func(ctx context.Context, eventName string, optionalData ...interface{}) {}

	cleanup := func() {
		utils.Client = origClient
		utils.Executor = origExecutor
	}
	return m, cleanup
}

func TestManager_Basic(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	m.CancelDownload("test")
}

func TestDeleteVersion(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

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
}

func TestGetInstalledVersionPaths(t *testing.T) {
	tempDir := t.TempDir()
	os.MkdirAll(filepath.Join(tempDir, "bin", "php", "php-8.1.0"), 0755)
	os.WriteFile(filepath.Join(tempDir, "bin", "php", "php-8.1.0", "php.exe"), []byte(""), 0644)

	paths := GetInstalledVersionPaths(tempDir, "php", "php.exe")
	if len(paths) == 0 {
		t.Error("Expected versions")
	}
}

func TestDiscovery(t *testing.T) {
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

	// Test HeidiSQL detection
	DetectHeidiSQLInstallationOverride = func() (string, string) {
		return "/usr/bin/heidisql.exe", ""
	}
	defer func() { DetectHeidiSQLInstallationOverride = nil }()

	// Create modules folders
	os.MkdirAll(filepath.Join(phpDir, "composer"), 0755)
	os.WriteFile(filepath.Join(phpDir, "composer.phar"), []byte(""), 0644)

	_ = GetLatestKnownVersions()
}

func TestDownloadAndExtract_Basic(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	tempDir := t.TempDir()
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	// Use a mock that provides a body to avoid errors in io.Copy
	utils.Client = &mockHTTPClientProgress{}

	task := DownloadTask{
		Name:      "TestPlugin",
		Version:   "1.0.0",
		URL:       "http://example.com/test.exe", // Ends in .exe, triggers handleInstaller
		Target:    "test/1.0.0",
		CheckFile: "test.exe",
	}

	err := m.DownloadAndExtract(context.Background(), task)
	if err != nil {
		t.Errorf("DownloadAndExtract failed: %v", err)
	}

	// Test download error path
	utils.Client = &mockHTTPClient{err: fmt.Errorf("download error")}
	err = m.DownloadAndExtract(context.Background(), task)
	if err == nil {
		t.Error("Expected download error")
	}
}

func TestGetInstalledVersionPaths_Extended(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	tempDir := t.TempDir()
	_ = m.GetInstalledVersionPaths("php", "php.exe")
	_ = tempDir
}

func TestModuleMethods(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	tempDir := t.TempDir()
	phpPath := filepath.Join(tempDir, "bin", "php", "current")
	os.MkdirAll(phpPath, 0755)
	os.WriteFile(filepath.Join(phpPath, "php.exe"), []byte(""), 0755)

	_ = m.InstallModule("Composer", phpPath, nil)
	_ = m.UninstallModule("Composer", phpPath)

	_ = m.InstallModule("Xdebug", phpPath, nil)
	_ = m.UninstallModule("Xdebug", phpPath)
}

func TestDownloadFileManual(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	utils.Client = &mockHTTPClient{content: "manual"}
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "manual.exe")
	err := m.DownloadFileManual(context.Background(), "http://example.com/manual.exe", dest, "Manual")
	if err != nil {
		t.Errorf("DownloadFileManual failed: %v", err)
	}
}

func TestHandleArchive_Mocked(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

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
		Name:      "TestZip",
		Version:   "1.0.0",
		URL:       "http://example.com/test.zip",
		Target:    "testzip/1.0.0",
		CheckFile: "sub/test.txt",
	}
	targetDir := filepath.Join(tempDir, "bin", task.Target)

	err := m.handleArchive(context.Background(), task, "dummy.zip", targetDir)
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
	err = m.handleArchive(context.Background(), task, "dummy.nupkg", targetDir+"_nupkg")
	if err != nil {
		t.Errorf("handleArchive nupkg failed: %v", err)
	}
}

func TestNewManager_Emit(t *testing.T) {
	mgr := NewManager(context.Background())
	// Directly test the emit logic without calling wruntime
	mgr.emit = func(ctx context.Context, eventName string, optionalData ...interface{}) {}
	mgr.emit(context.Background(), "test-event")
	mgr.emit(nil, "test-event")
}

func TestDownloadAndExtract_AlreadyInstalled(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	tempDir := t.TempDir()
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	task := DownloadTask{
		Name:      "PHP",
		Version:   "8.2.0",
		Target:    "php/php-8.2.0",
		CheckFile: "php.exe",
	}
	targetDir := filepath.Join(tempDir, "bin", task.Target)
	os.MkdirAll(targetDir, 0755)
	os.WriteFile(filepath.Join(targetDir, "php.exe"), []byte(""), 0644)

	err := m.DownloadAndExtract(context.Background(), task)
	if err != nil {
		t.Errorf("DownloadAndExtract should have skipped and returned nil, got %v", err)
	}

	// Test already installed branch with Apache special case
	task2 := DownloadTask{
		Name:      "Apache",
		Version:   "2.4.54",
		Target:    "apache/httpd-2.4.54",
		CheckFile: "bin/httpd.exe",
	}
	targetDir2 := filepath.Join(tempDir, "bin", task2.Target)
	os.MkdirAll(filepath.Join(targetDir2, "Apache24", "bin"), 0755)
	os.WriteFile(filepath.Join(targetDir2, "Apache24", "bin", "httpd.exe"), []byte(""), 0644)
	if !m.isAlreadyInstalled(context.Background(), task2, targetDir2) {
		t.Error("isAlreadyInstalled failed for Apache special case")
	}
}

func TestNewManager_ProdEmit(t *testing.T) {
	mgr := NewManager(context.Background())
	// Let's use a nil context to avoid hitting actual wruntime.EventsEmit, which panics in tests.
	mgr.emit(nil, "test-event-dummy")
}

func TestManager_PostProcessExtractionManual(t *testing.T) {
	m := &Manager{}
	tmpDir := t.TempDir()

	// Nested dir test
	extractTmp := filepath.Join(tmpDir, "extract")
	targetDir := filepath.Join(tmpDir, "target")
	os.MkdirAll(filepath.Join(extractTmp, "nested"), 0755)
	err := os.WriteFile(filepath.Join(extractTmp, "nested", "file.txt"), []byte("data"), 0644)
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	err = m.PostProcessExtractionManual(extractTmp, targetDir)
	if err != nil {
		t.Fatalf("PostProcessExtractionManual failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "file.txt")); err != nil {
		t.Errorf("Expected file.txt to exist in targetDir, error: %v", err)
	}
}

func TestDownloadAndExtract_Installer_CopyError(t *testing.T) {
	m, cleanup := setupTestManager(t)
	defer cleanup()

	tempDir := t.TempDir()
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	task := DownloadTask{
		Name:      "HeidiSQL",
		Version:   "12.0",
		URL:       "http://example.com/installer.exe",
		Target:    "heidisql/12.0",
		CheckFile: "heidisql.exe",
	}

	err := m.handleInstaller(context.Background(), task, filepath.Join(tempDir, "non_existent_installer.exe"), "/invalid_target_dir/")
	if err == nil {
		t.Error("Expected error from handleInstaller due to copy failure")
	}
}

func TestDiscovery_DetectModules_Extended(t *testing.T) {
	tmpDir := t.TempDir()
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	binDir := filepath.Join(tmpDir, "bin")
	phpDir := filepath.Join(binDir, "php", "current")
	os.MkdirAll(phpDir, 0755)
	os.WriteFile(filepath.Join(phpDir, "php.exe"), []byte(""), 0644)
	os.WriteFile(filepath.Join(phpDir, "composer.phar"), []byte(""), 0644)

	tasks := GetLatestKnownVersions()
	foundComposer := false
	for _, t := range tasks {
		if t.Name == "PHP" {
			for _, m := range t.Modules {
				if m.Name == "Composer" && m.IsInstalled {
					foundComposer = true
				}
			}
		}
	}
	if !foundComposer {
		t.Error("Expected Composer module to be detected as installed under php/current")
	}
}

func TestExtractor_Unzip_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	err := unzipFunc(context.Background(), "non_existent_archive.zip", tmpDir, "TestZip", nil)
	if err == nil {
		t.Error("Expected Unzip to fail for a non-existent archive")
	}
}
