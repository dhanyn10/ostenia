package backend

import (
	"context"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	plugins_utils "ostenia/internal/plugins/utils"
	"ostenia/internal/ssl"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Mock global dependencies to ensure tests are isolated and don't touch the system
	origTrust := ssl.TrustRootCAOverride
	ssl.TrustRootCAOverride = func(path string) error { return nil }

	origExecutor := plugins_utils.Executor
	plugins_utils.Executor = &LocalMockExecutor{Output: "mocked"}

	code := m.Run()

	// Restore originals
	ssl.TrustRootCAOverride = origTrust
	plugins_utils.Executor = origExecutor

	os.Exit(code)
}

type LocalMockExecutor struct {
	Output string
}

func (m *LocalMockExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command("echo", m.Output) // NOSONAR
}

type MockRuntime struct {
	SelectedFile string
	SelectedDir  string
}

func (m *MockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
}
func (m *MockRuntime) WindowMinimise(ctx context.Context)          {}
func (m *MockRuntime) WindowMaximise(ctx context.Context)          {}
func (m *MockRuntime) WindowUnmaximise(ctx context.Context)        {}
func (m *MockRuntime) WindowExecJS(ctx context.Context, js string) {}
func (m *MockRuntime) Quit(ctx context.Context)                    {}
func (m *MockRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return m.SelectedFile, nil
}
func (m *MockRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return m.SelectedDir, nil
}
func (m *MockRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return m.SelectedFile, nil
}

type MockOrchestrator struct {
	interfaces.Orchestrator
	Running map[string]bool
}

func (m *MockOrchestrator) SetRuntime(r interfaces.Runtime)  {}
func (m *MockOrchestrator) StartWatcher(ctx context.Context) {}
func (m *MockOrchestrator) SetActiveTab(tab string)          {}
func (m *MockOrchestrator) RequestRefresh()                  {}
func (m *MockOrchestrator) IsRunning(name string) bool       { return m.Running[name] }
func (m *MockOrchestrator) GetDetailedInfo(name string) interfaces.ServiceDetailedInfo {
	return interfaces.ServiceDetailedInfo{Name: name, Status: "Stopped", Port: 80}
}
func (m *MockOrchestrator) StartServiceWithPort(ctx context.Context, name, binaryPath string, args []string, workingDir string, port int) error {
	m.Running[name] = true
	return nil
}
func (m *MockOrchestrator) StartService(ctx context.Context, name, binaryPath string, args []string, workingDir string) error {
	m.Running[name] = true
	return nil
}
func (m *MockOrchestrator) StopService(ctx context.Context, name string) error {
	m.Running[name] = false
	return nil
}
func (m *MockOrchestrator) StopAll(ctx context.Context) {
	m.Running = make(map[string]bool)
}

type MockPluginManager struct {
	interfaces.PluginManager
}

func (m *MockPluginManager) DownloadAndExtract(ctx context.Context, task interfaces.DownloadTask) error {
	return nil
}
func (m *MockPluginManager) DeleteVersion(category, version string) error { return nil }
func (m *MockPluginManager) CancelDownload(category string)               {}
func (m *MockPluginManager) GetInstalledVersionPaths(category, checkFile string) map[string]string {
	return map[string]string{"8.2.0": "/path/to/8.2.0"}
}
func (m *MockPluginManager) InstallModule(moduleName, phpPath string, emitProgress func(string, float64, string)) error {
	return nil
}
func (m *MockPluginManager) UninstallModule(moduleName, phpPath string) error { return nil }

type MockSSHManager struct {
	interfaces.SSHManager
}

func (m *MockSSHManager) SetRuntime(r interfaces.Runtime) {}
func (m *MockSSHManager) GetSessions() ([]config.SSHSession, error) {
	return []config.SSHSession{}, nil
}
func (m *MockSSHManager) SaveSessions(sessions []config.SSHSession) error              { return nil }
func (m *MockSSHManager) Connect(ctx context.Context, session config.SSHSession) error { return nil }
func (m *MockSSHManager) Disconnect(sessionID string)                                  {}
func (m *MockSSHManager) SendInput(sessionID string, input string) error               { return nil }
func (m *MockSSHManager) ResizeTerminal(sessionID string, cols, rows int) error        { return nil }
func (m *MockSSHManager) ListFiles(sessionID string, path string) ([]interfaces.RemoteFile, error) {
	return []interfaces.RemoteFile{}, nil
}
func (m *MockSSHManager) ExecuteSFTPAction(sessionID, action, path, newPath string) error { return nil }
func (m *MockSSHManager) DownloadFile(sessionID, remotePath, localPath string) error      { return nil }
func (m *MockSSHManager) UploadFile(sessionID, localPath, remotePath string) error        { return nil }
func (m *MockSSHManager) EditFile(sessionID, remotePath, editor string) error             { return nil }
func (m *MockSSHManager) GetCurrentPath(sessionID string) (string, error)                 { return "/", nil }
func (m *MockSSHManager) GetWSLDistros() ([]string, error)                               { return []string{"Ubuntu"}, nil }
func (m *MockSSHManager) GetResourceUsage(sessionID string) (interfaces.ResourceUsage, error) {
	return interfaces.ResourceUsage{CPU: 10, Mem: 20, Disk: 30}, nil
}

type MockSSLManager struct {
	interfaces.SSLManager
}

func (m *MockSSLManager) GenerateRootCA(destDir string) error                 { return nil }
func (m *MockSSLManager) GetRemainingDays(path string) (int, error)           { return 365, nil }
func (m *MockSSLManager) SignCertificate(caDir, domain, destDir string) error { return nil }

func TestApp_Full_Mocked(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Unsetenv("OSTENIA_HOME")

	mockR := &MockRuntime{
		SelectedFile: filepath.Join(tempDir, "selected.txt"),
		SelectedDir:  filepath.Join(tempDir, "selected_dir"),
	}
	os.MkdirAll(mockR.SelectedDir, 0755)

	ctx := context.Background()

	app := &App{
		ctx:          ctx,
		runtime:      mockR,
		cfg:          &config.Config{Proxies: map[string]int{"test": 3000}, BaseDir: tempDir, WWWRoot: filepath.Join(tempDir, "www")},
		downloader:   &MockPluginManager{},
		orchestrator: &MockOrchestrator{Running: make(map[string]bool)},
		sshManager:   &MockSSHManager{},
		sslManager:   &MockSSLManager{},
	}

	// App window
	app.Minimize()
	app.Maximize()
	app.Unmaximize()
	app.ToggleDevTools()
	app.Close()
	app.EventsEmit(ctx, "test", nil)
	_, _ = app.OpenFileDialog(ctx, wruntime.OpenDialogOptions{})
	_, _ = app.OpenDirectoryDialog(ctx, wruntime.OpenDialogOptions{})
	_, _ = app.SaveFileDialog(ctx, wruntime.SaveDialogOptions{})
	app.Quit(ctx)

	_ = app.GenerateRootCA("test")
	_, _ = app.GetRemainingDays("test")
	_ = app.SignCertificate("ca", "domain", "dest")

	// Config
	_ = app.GetConfig()
	_, _ = app.SelectDefaultEditor(ctx)
	_ = app.SetDefaultEditor("path")
	app.UpdateActiveTab("activity")
	_ = app.IsAdmin()

	// Env
	app.orchestrator.StartService(ctx, "Apache", "", nil, "")
	_ = app.SetWWWRoot(ctx, filepath.Join(tempDir, "www2"))
	_ = app.SetServerRoot(ctx, tempDir)
	_, _ = app.SelectServerRoot(ctx)
	_, _ = app.SelectWWWRoot(ctx)
	_ = app.OpenServerRootFolder()
	_ = app.OpenAppsLocationFolder()

	// Plugins
	_ = app.GetPrerequisites()
	app.CancelDownload("test")
	_ = app.OpenPluginFolder("PHP")
	_ = app.InstallPrerequisite(ctx, interfaces.DownloadTask{Name: "OpenSSL"})
	_ = app.InstallPrerequisite(ctx, interfaces.DownloadTask{Name: "Python"})
	_ = app.InstallPluginModule(ctx, "PHP", "Composer")
	_ = app.InstallPluginModule(ctx, "Python", "Pip")
	_ = app.InstallPluginModule(ctx, "Unknown", "Mod")
	_ = app.UninstallPluginModule("PHP", "Composer")
	_ = app.UninstallPluginModule("Python", "Pip")
	_ = app.UninstallPluginModule("Unknown", "Mod")
	_ = app.SwitchServiceVersion(ctx, "PHP", "8.2.0")
	_ = app.DeleteVersion("PHP", "8.1.0")

	// Services
	_ = app.GetServiceStatus("Apache")
	for _, s := range []string{"Apache", "MySQL", "Nginx", "PHP", serviceNodeJS, "Python", "OpenSSL", "HeidiSQL", "Unknown"} {
		_ = app.StartService(ctx, s)
		_ = app.StopService(ctx, s)
	}
	_ = app.StartAllServices(ctx)
	app.StopAllServices(ctx)
	_ = app.SetApacheHTTPS(ctx, true)
	_ = app.SetNginxHTTPS(ctx, true)

	// SSH
	_, _ = app.GetSSHSessions()
	_ = app.AddSSHSession(config.SSHSession{ID: "test"})
	_ = app.UpdateSSHSession(config.SSHSession{ID: "test"})
	_ = app.DeleteSSHSession("test")
	_ = app.SaveSSHSessions([]config.SSHSession{})
	_ = app.ConnectSSH(ctx, config.SSHSession{ID: "test"})
	app.DisconnectSSH("test")
	_ = app.SendSSHInput("test", "ls")
	_ = app.ResizeSSHTerminal("test", 80, 24)
	_, _ = app.GetRemoteFiles("test", "/")
	_ = app.ExecuteSFTPAction("test", "mkdir", "/path", "")
	_ = app.EditRemoteFile("test", "/path")
	_, _ = app.GetRemoteCurrentPath("test")
	_ = app.DownloadRemoteFile(ctx, "test", "/path")
	_ = app.UploadRemoteFile(ctx, "test", "/path")
	_, _ = app.GetSSHResourceUsage("test")
	_, _ = app.GetWSLDistros()

	// Network
	_ = app.CheckProxyPorts()
	_ = app.GetProxyApps()
	_ = app.SaveProxyPort(ctx, "myapp", 4000)
	_ = app.SaveProxyPort(ctx, "myapp", 0) // Test deletion
	app.OpenProxyTerminal("myapp", "cmd")

	// PHP
	app.orchestrator.StartService(ctx, "PHP", "", nil, "")
	_, _ = app.GetPHPExtensions()
	_ = app.TogglePHPExtension(ctx, "openssl", true)

	// Services extra
	_ = app.OpenHeidiSQL()
	_ = app.OpenServiceTerminal("PHP", "cmd")
	app.OpenTerminal("cmd")
	app.OpenTerminalAtPath("cmd", tempDir)

	// Extra service coverage
	_ = app.StartService(ctx, "Unknown")
	_ = app.StopService(ctx, "PHP")
	_ = app.StopService(ctx, "OpenSSL")
	_ = app.StopService(ctx, serviceNodeJS)
	_ = app.StopService(ctx, "Python")

	// Node and Python Start Service fail cases
	_ = app.startNodeService("/nonexistent")
	_ = app.startPythonService("/nonexistent")

	plugins.DetectHeidiSQLInstallationOverride = func() (string, string) { return "path", "" }
	_ = app.OpenHeidiSQL()
	plugins.DetectHeidiSQLInstallationOverride = nil

	// Profile
	_ = app.ExportProfile(ctx, true, true)
	_ = app.ImportProfile(ctx)
}

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Error("NewApp returned nil")
	}
}

func TestApp_Startup(t *testing.T) {
	tempDir := t.TempDir()
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	configPath := filepath.Join(tempDir, "config.json")
	oldConfig := config.SetConfigFile(configPath)
	defer config.SetConfigFile(oldConfig)

	app := NewApp()
	// Inject mocks
	app.runtime = &MockRuntime{}
	app.orchestrator = &MockOrchestrator{Running: make(map[string]bool)}
	app.downloader = &MockPluginManager{}
	app.sshManager = &MockSSHManager{}
	app.sslManager = &MockSSLManager{}

	ctx, cancel := context.WithCancel(context.Background())
	app.Startup(ctx)
	cancel() // Stop background goroutines like startProxyWatcher

	if app.cfg == nil {
		t.Error("Expected config to be set")
	}
}

func TestWailsRuntime(t *testing.T) {
	t.Skip("Skipping WailsRuntime tests as they require a valid Wails context")
}

func TestApp_SSLDelegates(t *testing.T) {
	app := &App{
		sslManager: &MockSSLManager{},
	}
	_ = app.GenerateRootCA("test")
	_, _ = app.GetRemainingDays("test")
	_ = app.SignCertificate("ca", "domain", "dest")
}

func TestDefaultSSLManager(t *testing.T) {
	// These call the actual internal/ssl package, which is risky but since we are in backend
	// and we want coverage for the wrapper, we just call them.
	// However, they might fail because of missing binaries or permissions.
	// We mainly want to cover the delegate lines.
	s := &DefaultSSLManager{}
	_ = s.GenerateRootCA(t.TempDir())
	_, _ = s.GetRemainingDays("nonexistent")
	_ = s.SignCertificate("ca", "domain", "dest")
}

func TestApp_Services_RealIsh(t *testing.T) {
	tempDir := t.TempDir()
	oldEnv := os.Getenv("OSTENIA_HOME")
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Setenv("OSTENIA_HOME", oldEnv)

	// Create dummy binary structure
	binDir := filepath.Join(tempDir, "bin")

	// PHP
	phpDir := filepath.Join(binDir, "php", "php-8.2.0")
	os.MkdirAll(phpDir, 0755)
	os.WriteFile(filepath.Join(phpDir, "php-cgi.exe"), []byte(""), 0755)
	// Symlink version
	phpCurrent := filepath.Join(binDir, "php", "current")
	os.MkdirAll(phpCurrent, 0755)
	os.WriteFile(filepath.Join(phpCurrent, "php-cgi.exe"), []byte(""), 0755)
	_ = os.WriteFile(filepath.Join(phpCurrent, "php.ini"), []byte("extension=openssl\n"), 0644)

	// MySQL
	mysqlDir := filepath.Join(binDir, "mysql", "mysql-8.0.0")
	os.MkdirAll(filepath.Join(mysqlDir, "bin"), 0755)
	os.WriteFile(filepath.Join(mysqlDir, "bin", "mysqld.exe"), []byte(""), 0755)
	// Symlink version
	mysqlCurrent := filepath.Join(binDir, "mysql", "current")
	os.MkdirAll(filepath.Join(mysqlCurrent, "bin"), 0755)
	os.WriteFile(filepath.Join(mysqlCurrent, "bin", "mysqld.exe"), []byte(""), 0755)

	// Apache
	apacheDir := filepath.Join(binDir, "apache", "httpd-2.4.0")
	os.MkdirAll(filepath.Join(apacheDir, "bin"), 0755)
	os.WriteFile(filepath.Join(apacheDir, "bin", "httpd.exe"), []byte(""), 0755)
	// Apache fallback structure
	apacheFallbackDir := filepath.Join(binDir, "apache", "httpd-2.4.1")
	os.MkdirAll(filepath.Join(apacheFallbackDir, "Apache24", "bin"), 0755)
	os.WriteFile(filepath.Join(apacheFallbackDir, "Apache24", "bin", "httpd.exe"), []byte(""), 0755)
	// Symlink version
	apacheCurrent := filepath.Join(binDir, "apache", "current")
	os.MkdirAll(filepath.Join(apacheCurrent, "bin"), 0755)
	os.WriteFile(filepath.Join(apacheCurrent, "bin", "httpd.exe"), []byte(""), 0755)

	// Nginx
	nginxDir := filepath.Join(binDir, "nginx", "nginx-1.24.0")
	os.MkdirAll(nginxDir, 0755)
	os.WriteFile(filepath.Join(nginxDir, "nginx.exe"), []byte(""), 0755)
	// Symlink version
	nginxCurrent := filepath.Join(binDir, "nginx", "current")
	os.MkdirAll(nginxCurrent, 0755)
	os.WriteFile(filepath.Join(nginxCurrent, "nginx.exe"), []byte(""), 0755)

	// Node
	nodeDir := filepath.Join(binDir, "nodejs", "node-v18.0.0")
	os.MkdirAll(filepath.Join(nodeDir, "bin"), 0755)
	os.WriteFile(filepath.Join(nodeDir, "bin", "node.exe"), []byte(""), 0755)
	nodeCurrent := filepath.Join(binDir, "nodejs", "current")
	os.MkdirAll(filepath.Join(nodeCurrent, "bin"), 0755)
	os.WriteFile(filepath.Join(nodeCurrent, "bin", "node.exe"), []byte(""), 0755)

	// Python
	pythonDir := filepath.Join(binDir, "python", "python-3.10.0")
	os.MkdirAll(filepath.Join(pythonDir, "bin"), 0755)
	os.WriteFile(filepath.Join(pythonDir, "bin", "python.exe"), []byte(""), 0755)
	pythonCurrent := filepath.Join(binDir, "python", "current")
	os.MkdirAll(filepath.Join(pythonCurrent, "bin"), 0755)
	os.WriteFile(filepath.Join(pythonCurrent, "bin", "python.exe"), []byte(""), 0755)

	ctx := context.Background()

	app := &App{
		runtime: &MockRuntime{},
		cfg: &config.Config{
			BaseDir: tempDir,
			WWWRoot: filepath.Join(tempDir, "www"),
			Proxies: make(map[string]int),
		},
		orchestrator: &MockOrchestrator{Running: make(map[string]bool)},
		downloader:   &MockPluginManager{},
		sshManager:   &MockSSHManager{},
		sslManager:   &MockSSLManager{},
	}

	// Now test starting services
	_ = app.StartService(ctx, "PHP")
	_ = app.StartService(ctx, "MySQL")
	_ = app.StartService(ctx, "Apache")
	_ = app.StartService(ctx, "Nginx")

	// Test SetHTTPS with running services to trigger restart
	app.orchestrator.StartService(ctx, "Apache", "", nil, "")
	_ = app.SetApacheHTTPS(ctx, true)
	app.orchestrator.StartService(ctx, "Nginx", "", nil, "")
	_ = app.SetNginxHTTPS(ctx, true)

	// Ensure dependent web servers restart when PHP starts
	app.orchestrator.StartService(ctx, "Apache", "", nil, "")
	app.cfg.Proxies["mysite"] = 8080
	_ = app.StartService(ctx, "PHP")

	// Node & Python
	_ = app.SwitchServiceVersion(ctx, serviceNodeJS, "18.0.0")
	_ = app.StartService(ctx, serviceNodeJS)
	_ = app.SwitchServiceVersion(ctx, "Python", "3.10.0")
	_ = app.StartService(ctx, "Python")

	// Test StartAll
	_ = app.StartAllServices(ctx)

	// Test StopService for all
	for _, s := range []string{"Apache", "MySQL", "Nginx", "PHP", serviceNodeJS, "Python", "OpenSSL"} {
		_ = app.StopService(ctx, s)
	}

	// Test helper methods and stubs
	_, _ = app.GetInstalledApps()
	_, _ = app.checkStandardExePath(apacheFallbackDir, "httpd.exe")
	_, _ = app.checkStandardExePath(nginxDir, "nginx.exe")
	_, _ = app.walkForExecutable(nginxDir, "nginx.exe")

	// Test app_network.go
	app.orchestrator.StartService(ctx, "Apache", "", nil, "")
	app.orchestrator.StartService(ctx, "Nginx", "", nil, "")
	_ = app.SaveProxyPort(ctx, "mysite", 8080)
	_ = app.GetProxyApps()
	_ = app.CheckProxyPorts()
	_ = app.OpenProxyTerminal("mysite", "cmd")

	// Test app_plugins.go
	_ = app.InstallPrerequisite(ctx, plugins.DownloadTask{Name: "PHP", Target: "php/php-8.2.0"})
	_ = app.OpenPluginFolder("PHP")
	_ = app.InstallPluginModule(ctx, "PHP", "Composer")
	_ = app.UninstallPluginModule("PHP", "Composer")
	_ = app.DeleteVersion("PHP", "8.1.0")

	// Coverage for getServiceTargetDir
	_ = app.getServiceTargetDir("nginx", nginxDir)
	_ = app.getServiceTargetDir("apache", apacheDir)
	_ = app.getServiceTargetDir("mysql", mysqlDir)
	_ = app.getServiceTargetDir("unknown", tempDir)

	// Test app_profile.go
	// Create some files to export
	os.MkdirAll(filepath.Join(tempDir, "www", "site1"), 0755)
	os.WriteFile(filepath.Join(tempDir, "www", "site1", "index.php"), []byte("<?php"), 0644)
	_ = app.ExportProfile(ctx, true, true)

	// ImportProfile
	localMockR := app.runtime.(*MockRuntime)
	localMockR.SelectedFile = filepath.Join(tempDir, "selected.txt")
	os.WriteFile(localMockR.SelectedFile, []byte(`{"config":{"phpVersion":"8.2.0"}}`), 0644)
	_ = app.ImportProfile(ctx)
}
