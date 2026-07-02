package backend

import (
	"context"
	"ostenia/internal/config"
	"ostenia/internal/backend/interfaces"
	"testing"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
    "os"
    "path/filepath"
)

type MockRuntime struct {
    SelectedFile string
    SelectedDir  string
}

func (m *MockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {}
func (m *MockRuntime) WindowMinimise(ctx context.Context)   {}
func (m *MockRuntime) WindowMaximise(ctx context.Context)   {}
func (m *MockRuntime) WindowUnmaximise(ctx context.Context) {}
func (m *MockRuntime) WindowExecJS(ctx context.Context, js string) {}
func (m *MockRuntime) Quit(ctx context.Context)             {}
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
func (m *MockOrchestrator) SetRuntime(r interfaces.Runtime) {}
func (m *MockOrchestrator) SetActiveTab(tab string)        {}
func (m *MockOrchestrator) RequestRefresh()                {}
func (m *MockOrchestrator) StartWatcher()                  {}
func (m *MockOrchestrator) IsRunning(name string) bool     { return m.Running[name] }
func (m *MockOrchestrator) GetDetailedInfo(name string) interfaces.ServiceDetailedInfo {
    return interfaces.ServiceDetailedInfo{Name: name, Status: "Stopped", Port: 80}
}
func (m *MockOrchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
    m.Running[name] = true
    return nil
}
func (m *MockOrchestrator) StartService(name string, binaryPath string, args []string, workingDir string) error {
    m.Running[name] = true
    return nil
}
func (m *MockOrchestrator) StopService(name string) error {
    m.Running[name] = false
    return nil
}
func (m *MockOrchestrator) StopAll() {
    m.Running = make(map[string]bool)
}

type MockPluginManager struct {
    interfaces.PluginManager
}
func (m *MockPluginManager) DownloadAndExtract(task interfaces.DownloadTask) error { return nil }
func (m *MockPluginManager) DeleteVersion(category, version string) error { return nil }
func (m *MockPluginManager) CancelDownload(category string) {}
func (m *MockPluginManager) GetInstalledVersionPaths(category, checkFile string) map[string]string {
    return map[string]string{"8.2.0": "/path/to/8.2.0"}
}
func (m *MockPluginManager) InstallModule(moduleName string, phpPath string, emitProgress func(string, float64, string)) error { return nil }
func (m *MockPluginManager) UninstallModule(moduleName string, phpPath string) error { return nil }

type MockSSHManager struct {
    interfaces.SSHManager
}
func (m *MockSSHManager) GetSessions() ([]config.SSHSession, error) { return []config.SSHSession{}, nil }
func (m *MockSSHManager) SaveSessions(sessions []config.SSHSession) error { return nil }
func (m *MockSSHManager) Connect(session config.SSHSession) error { return nil }
func (m *MockSSHManager) Disconnect(sessionID string) {}
func (m *MockSSHManager) SendInput(sessionID string, input string) error { return nil }
func (m *MockSSHManager) ResizeTerminal(sessionID string, cols, rows int) error { return nil }
func (m *MockSSHManager) ListFiles(sessionID string, path string) ([]interfaces.RemoteFile, error) { return []interfaces.RemoteFile{}, nil }
func (m *MockSSHManager) ExecuteSFTPAction(sessionID string, action, path, newPath string) error { return nil }
func (m *MockSSHManager) DownloadFile(sessionID string, remotePath, localPath string) error { return nil }
func (m *MockSSHManager) UploadFile(sessionID string, localPath, remotePath string) error { return nil }
func (m *MockSSHManager) EditFile(sessionID string, remotePath, editor string) error { return nil }
func (m *MockSSHManager) GetCurrentPath(sessionID string) (string, error) { return "/", nil }

type MockSSLManager struct {
    interfaces.SSLManager
}
func (m *MockSSLManager) GenerateRootCA(destDir string) error { return nil }
func (m *MockSSLManager) GetRemainingDays(path string) (int, error) { return 365, nil }
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

	app := &App{
		runtime:      mockR,
		ctx:          context.Background(),
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

    // Config
	_ = app.GetConfig()
	_, _ = app.SelectDefaultEditor()
	_ = app.SetDefaultEditor("path")
	app.UpdateActiveTab("activity")
	_ = app.IsAdmin()

    // Env
	_ = app.SetWWWRoot(filepath.Join(tempDir, "www2"))
	_ = app.SetServerRoot(tempDir)
	_, _ = app.SelectServerRoot()
	_, _ = app.SelectWWWRoot()
	_ = app.OpenServerRootFolder()
	_ = app.OpenAppsLocationFolder()

    // Plugins
	_ = app.GetPrerequisites()
	app.CancelDownload("test")
    _ = app.InstallPluginModule("PHP", "Composer")
    _ = app.UninstallPluginModule("PHP", "Composer")
    _ = app.SwitchServiceVersion("PHP", "8.2.0")
    _ = app.DeleteVersion("PHP", "8.1.0")

    // Services
	_ = app.GetServiceStatus("Apache")
	for _, s := range []string{"Apache", "MySQL", "Nginx", "PHP", "Node.js", "Python", "OpenSSL", "HeidiSQL", "Unknown"} {
		_ = app.StartService(s)
		_ = app.StopService(s)
	}
	_ = app.StartAllServices()
	app.StopAllServices()
	_ = app.SetApacheHTTPS(true)
	_ = app.SetNginxHTTPS(true)

    // SSH
	_, _ = app.GetSSHSessions()
	_ = app.AddSSHSession(config.SSHSession{ID: "test"})
	_ = app.UpdateSSHSession(config.SSHSession{ID: "test"})
	_ = app.DeleteSSHSession("test")
	_ = app.SaveSSHSessions([]config.SSHSession{})
    _ = app.ConnectSSH(config.SSHSession{ID: "test"})
    app.DisconnectSSH("test")
    _ = app.SendSSHInput("test", "ls")
    _ = app.ResizeSSHTerminal("test", 80, 24)
    _, _ = app.GetRemoteFiles("test", "/")
    _ = app.ExecuteSFTPAction("test", "mkdir", "/path", "")
    _ = app.EditRemoteFile("test", "/path")
    _, _ = app.GetRemoteCurrentPath("test")
    _ = app.DownloadRemoteFile("test", "/path")
	_ = app.UploadRemoteFile("test", "/path")

    // Network
    _ = app.CheckProxyPorts()
    _ = app.GetProxyApps()
    _ = app.SaveProxyPort("myapp", 4000)

    // PHP
    _, _ = app.GetPHPExtensions()
    _ = app.TogglePHPExtension("openssl", true)

    // Profile
    _ = app.ExportProfile(true, true)
    _ = app.ImportProfile()
}

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil { t.Error("NewApp returned nil") }
}
