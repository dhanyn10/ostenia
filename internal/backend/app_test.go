package backend

import (
	"context"
	"ostenia/internal/config"
	"ostenia/internal/backend/interfaces"
	"testing"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
    "os"
    "path/filepath"
    "github.com/stretchr/testify/assert"
)

type MockRuntime struct {
    SelectedFile string
    SelectedDir  string
    EventsEmitCalled bool
    WindowMinimiseCalled bool
    WindowMaximiseCalled bool
    WindowUnmaximiseCalled bool
    WindowExecJSCalled bool
    QuitCalled bool
    OpenFileDialogCalled bool
    OpenDirectoryDialogCalled bool
    SaveFileDialogCalled bool
    ToggleDevToolsCalled bool
    CloseCalled bool
}

func (m *MockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) { m.EventsEmitCalled = true }
func (m *MockRuntime) WindowMinimise(ctx context.Context)   { m.WindowMinimiseCalled = true }
func (m *MockRuntime) WindowMaximise(ctx context.Context)   { m.WindowMaximiseCalled = true }
func (m *MockRuntime) WindowUnmaximise(ctx context.Context) { m.WindowUnmaximiseCalled = true }
func (m *MockRuntime) WindowExecJS(ctx context.Context, js string) { m.WindowExecJSCalled = true }
func (m *MockRuntime) Quit(ctx context.Context)             { m.QuitCalled = true }
func (m *MockRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
    m.OpenFileDialogCalled = true
	return m.SelectedFile, nil
}
func (m *MockRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
    m.OpenDirectoryDialogCalled = true
	return m.SelectedDir, nil
}
func (m *MockRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
    m.SaveFileDialogCalled = true
	return m.SelectedFile, nil
}
// Assuming these are part of the runtime interface or directly called on Wails runtime
func (m *MockRuntime) ToggleDevTools(ctx context.Context) { m.ToggleDevToolsCalled = true }
func (m *MockRuntime) Close(ctx context.Context)          { m.CloseCalled = true }


type MockOrchestrator struct {
    interfaces.Orchestrator
    Running map[string]bool
    StartServiceCalled map[string]bool
    StopServiceCalled map[string]bool
    StopAllCalled bool
    SetActiveTabCalled bool
    GetServiceStatusCalled map[string]bool
    SetApacheHTTPSCalled bool
    SetNginxHTTPSCalled bool
}
func (m *MockOrchestrator) SetRuntime(r interfaces.Runtime) {}
func (m *MockOrchestrator) StartWatcher()                  {}
func (m *MockOrchestrator) SetActiveTab(tab string)        { m.SetActiveTabCalled = true }
func (m *MockOrchestrator) RequestRefresh()                {}
func (m *MockOrchestrator) IsRunning(name string) bool     { return m.Running[name] }
func (m *MockOrchestrator) GetDetailedInfo(name string) interfaces.ServiceDetailedInfo {
    if m.GetServiceStatusCalled == nil { m.GetServiceStatusCalled = make(map[string]bool) }
    m.GetServiceStatusCalled[name] = true
    return interfaces.ServiceDetailedInfo{Name: name, Status: "Stopped", Port: 80}
}
func (m *MockOrchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
    if m.StartServiceCalled == nil { m.StartServiceCalled = make(map[string]bool) }
    m.StartServiceCalled[name] = true
    m.Running[name] = true
    return nil
}
func (m *MockOrchestrator) StartService(name string, binaryPath string, args []string, workingDir string) error {
    if m.StartServiceCalled == nil { m.StartServiceCalled = make(map[string]bool) }
    m.StartServiceCalled[name] = true
    m.Running[name] = true
    return nil
}
func (m *MockOrchestrator) StopService(name string) error {
    if m.StopServiceCalled == nil { m.StopServiceCalled = make(map[string]bool) }
    m.StopServiceCalled[name] = true
    m.Running[name] = false
    return nil
}
func (m *MockOrchestrator) StopAll() {
    m.StopAllCalled = true
    m.Running = make(map[string]bool)
}
func (m *MockOrchestrator) SetApacheHTTPS(enable bool) error { m.SetApacheHTTPSCalled = true; return nil }
func (m *MockOrchestrator) SetNginxHTTPS(enable bool) error { m.SetNginxHTTPSCalled = true; return nil }


type MockPluginManager struct {
    interfaces.PluginManager
    DownloadAndExtractCalled bool
    DeleteVersionCalled bool
    CancelDownloadCalled bool
    InstallModuleCalled bool
    UninstallModuleCalled bool
    GetInstalledVersionPathsCalled bool
    GetPrerequisitesCalled bool
    OpenPluginFolderCalled bool
    SwitchServiceVersionCalled bool
}
func (m *MockPluginManager) DownloadAndExtract(task interfaces.DownloadTask) error { m.DownloadAndExtractCalled = true; return nil }
func (m *MockPluginManager) DeleteVersion(category, version string) error { m.DeleteVersionCalled = true; return nil }
func (m *MockPluginManager) CancelDownload(category string) { m.CancelDownloadCalled = true }
func (m *MockPluginManager) GetInstalledVersionPaths(category, checkFile string) map[string]string {
    m.GetInstalledVersionPathsCalled = true
    return map[string]string{"8.2.0": "/path/to/8.2.0"}
}
func (m *MockPluginManager) InstallModule(moduleName string, phpPath string, emitProgress func(string, float64, string)) error { m.InstallModuleCalled = true; return nil }
func (m *MockPluginManager) UninstallModule(moduleName string, phpPath string) error { m.UninstallModuleCalled = true; return nil }
func (m *MockPluginManager) GetPrerequisites() ([]interfaces.DownloadTask, error) { m.GetPrerequisitesCalled = true; return []interfaces.DownloadTask{}, nil }
func (m *MockPluginManager) OpenPluginFolder(category string) error { m.OpenPluginFolderCalled = true; return nil }
func (m *MockPluginManager) SwitchServiceVersion(category, version string) error { m.SwitchServiceVersionCalled = true; return nil }


type MockSSHManager struct {
    interfaces.SSHManager
    GetSessionsCalled bool
    SaveSessionsCalled bool
    ConnectCalled bool
    DisconnectCalled bool
    SendInputCalled bool
    ResizeTerminalCalled bool
    ListFilesCalled bool
    ExecuteSFTPActionCalled bool
    DownloadFileCalled bool
    UploadFileCalled bool
    EditFileCalled bool
    GetCurrentPathCalled bool
    AddSSHSessionCalled bool
    UpdateSSHSessionCalled bool
    DeleteSSHSessionCalled bool
}
func (m *MockSSHManager) GetSessions() ([]config.SSHSession, error) { m.GetSessionsCalled = true; return []config.SSHSession{}, nil }
func (m *MockSSHManager) SaveSessions(sessions []config.SSHSession) error { m.SaveSessionsCalled = true; return nil }
func (m *MockSSHManager) Connect(session config.SSHSession) error { m.ConnectCalled = true; return nil }
func (m *MockSSHManager) Disconnect(sessionID string) { m.DisconnectCalled = true }
func (m *MockSSHManager) SendInput(sessionID string, input string) error { m.SendInputCalled = true; return nil }
func (m *MockSSHManager) ResizeTerminal(sessionID string, cols, rows int) error { m.ResizeTerminalCalled = true; return nil }
func (m *MockSSHManager) ListFiles(sessionID string, path string) ([]interfaces.RemoteFile, error) { m.ListFilesCalled = true; return []interfaces.RemoteFile{}, nil }
func (m *MockSSHManager) ExecuteSFTPAction(sessionID string, action, path, newPath string) error { m.ExecuteSFTPActionCalled = true; return nil }
func (m *MockSSHManager) DownloadFile(sessionID string, remotePath, localPath string) error { m.DownloadFileCalled = true; return nil }
func (m *MockSSHManager) UploadFile(sessionID string, localPath, remotePath string) error { m.UploadFileCalled = true; return nil }
func (m *MockSSHManager) EditFile(sessionID string, remotePath, editor string) error { m.EditFileCalled = true; return nil }
func (m *MockSSHManager) GetCurrentPath(sessionID string) (string, error) { m.GetCurrentPathCalled = true; return "/", nil }
func (m *MockSSHManager) AddSSHSession(session config.SSHSession) error { m.AddSSHSessionCalled = true; return nil }
func (m *MockSSHManager) UpdateSSHSession(session config.SSHSession) error { m.UpdateSSHSessionCalled = true; return nil }
func (m *MockSSHManager) DeleteSSHSession(sessionID string) error { m.DeleteSSHSessionCalled = true; return nil }


type MockSSLManager struct {
    interfaces.SSLManager
    GenerateRootCACalled bool
    GetRemainingDaysCalled bool
    SignCertificateCalled bool
}
func (m *MockSSLManager) GenerateRootCA(destDir string) error { m.GenerateRootCACalled = true; return nil }
func (m *MockSSLManager) GetRemainingDays(path string) (int, error) { m.GetRemainingDaysCalled = true; return 365, nil }
func (m *MockSSLManager) SignCertificate(caDir, domain, destDir string) error { m.SignCertificateCalled = true; return nil }

type MockSymlinkManager struct {
    OpenFolderCalled bool
}
func (m *MockSymlinkManager) OpenFolder(path string) error { m.OpenFolderCalled = true; return nil }

func setupApp(t *testing.T) (*App, *MockRuntime, *MockOrchestrator, *MockPluginManager, *MockSSHManager, *MockSSLManager, *MockSymlinkManager, string) {
    tempDir := t.TempDir()
    os.Setenv("OSTENIA_HOME", tempDir)

	mockR := &MockRuntime{
        SelectedFile: filepath.Join(tempDir, "selected.txt"),
        SelectedDir:  filepath.Join(tempDir, "selected_dir"),
    }
    os.MkdirAll(mockR.SelectedDir, 0755)

    mockOrchestrator := &MockOrchestrator{Running: make(map[string]bool)}
    mockPluginManager := &MockPluginManager{}
    mockSSHManager := &MockSSHManager{}
    mockSSLManager := &MockSSLManager{}
    mockSymlinkManager := &MockSymlinkManager{}

	app := &App{
		runtime:      mockR,
		ctx:          context.Background(),
		cfg:          &config.Config{Proxies: map[string]int{"test": 3000}, BaseDir: tempDir, WWWRoot: filepath.Join(tempDir, "www")},
		downloader:   mockPluginManager,
		orchestrator: mockOrchestrator,
		sshManager:   mockSSHManager,
		sslManager:   mockSSLManager,
        symlinkMgr:   mockSymlinkManager,
	}
    return app, mockR, mockOrchestrator, mockPluginManager, mockSSHManager, mockSSLManager, mockSymlinkManager, tempDir
}

func teardownApp(tempDir string) {
    os.Unsetenv("OSTENIA_HOME")
    // t.TempDir() handles cleanup, so os.RemoveAll is not strictly necessary here
}

func TestApp_WindowFunctions(t *testing.T) {
    app, mockR, _, _, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    app.Minimize()
    assert.True(t, mockR.WindowMinimiseCalled, "WindowMinimise should have been called")

    app.Maximize()
    assert.True(t, mockR.WindowMaximiseCalled, "WindowMaximise should have been called")

    app.Unmaximize()
    assert.True(t, mockR.WindowUnmaximiseCalled, "WindowUnmaximise should have been called")

    app.ToggleDevTools()
    assert.True(t, mockR.ToggleDevToolsCalled, "ToggleDevTools should have been called")

    app.Close()
    assert.True(t, mockR.CloseCalled, "Close should have been called")

    app.EventsEmit(context.Background(), "test", nil)
    assert.True(t, mockR.EventsEmitCalled, "EventsEmit should have been called")

    _, _ = app.OpenFileDialog(context.Background(), wruntime.OpenDialogOptions{})
    assert.True(t, mockR.OpenFileDialogCalled, "OpenFileDialog should have been called")

    _, _ = app.OpenDirectoryDialog(context.Background(), wruntime.OpenDialogOptions{})
    assert.True(t, mockR.OpenDirectoryDialogCalled, "OpenDirectoryDialog should have been called")

    _, _ = app.SaveFileDialog(context.Background(), wruntime.SaveDialogOptions{})
    assert.True(t, mockR.SaveFileDialogCalled, "SaveFileDialog should have been called")

    app.Quit(context.Background())
    assert.True(t, mockR.QuitCalled, "Quit should have been called")
}

func TestApp_SSLFunctions(t *testing.T) {
    app, _, _, _, _, mockSSL, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    _ = app.GenerateRootCA("test")
    assert.True(t, mockSSL.GenerateRootCACalled, "GenerateRootCA should have been called")

    _, _ = app.GetRemainingDays("test")
    assert.True(t, mockSSL.GetRemainingDaysCalled, "GetRemainingDays should have been called")

    _ = app.SignCertificate("ca", "domain", "dest")
    assert.True(t, mockSSL.SignCertificateCalled, "SignCertificate should have been called")
}

func TestApp_ServiceFunctions(t *testing.T) {
    app, _, mockOrchestrator, _, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    // Test GetServiceStatus
    _ = app.GetServiceStatus("Apache")
    assert.True(t, mockOrchestrator.GetServiceStatusCalled["Apache"], "GetServiceStatus should have been called for Apache")

    // Test StartService and StopService for a specific service
    serviceName := "Apache"
    _ = app.StartService(serviceName)
    assert.True(t, mockOrchestrator.StartServiceCalled[serviceName], "StartService should have been called for %s", serviceName)
    assert.True(t, mockOrchestrator.IsRunning(serviceName), "%s should be running", serviceName)

    _ = app.StopService(serviceName)
    assert.True(t, mockOrchestrator.StopServiceCalled[serviceName], "StopService should have been called for %s", serviceName)
    assert.False(t, mockOrchestrator.IsRunning(serviceName), "%s should not be running", serviceName)

    // Test StartAllServices and StopAllServices
    mockOrchestrator.StartServiceCalled = make(map[string]bool) // Reset for this test
    mockOrchestrator.StopServiceCalled = make(map[string]bool)  // Reset for this test
    mockOrchestrator.StopAllCalled = false

    _ = app.StartAllServices()
    // We don't have direct access to the list of services App starts,
    // but we can check if some known services were attempted to start.
    // This assumes StartAllServices iterates over a predefined list.
    // For a more robust test, we'd need to mock the service discovery.
    assert.True(t, mockOrchestrator.StartServiceCalled["Apache"], "StartAllServices should attempt to start Apache")
    assert.True(t, mockOrchestrator.StartServiceCalled["MySQL"], "StartAllServices should attempt to start MySQL")

    app.StopAllServices()
    assert.True(t, mockOrchestrator.StopAllCalled, "StopAllServices should have been called")
    assert.False(t, mockOrchestrator.IsRunning("Apache"), "Apache should not be running after StopAllServices")

    _ = app.SetApacheHTTPS(true)
    assert.True(t, mockOrchestrator.SetApacheHTTPSCalled, "SetApacheHTTPS should have been called")

    _ = app.SetNginxHTTPS(true)
    assert.True(t, mockOrchestrator.SetNginxHTTPSCalled, "SetNginxHTTPS should have been called")
}

func TestApp_PluginFunctions(t *testing.T) {
    app, _, _, mockPlugin, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    _ = app.GetPrerequisites()
    assert.True(t, mockPlugin.GetPrerequisitesCalled, "GetPrerequisites should have been called")

    _ = app.OpenPluginFolder("PHP")
    assert.True(t, mockPlugin.OpenPluginFolderCalled, "OpenPluginFolder should have been called")

    _ = app.InstallPrerequisite(interfaces.DownloadTask{Name: "OpenSSL"})
    assert.True(t, mockPlugin.DownloadAndExtractCalled, "DownloadAndExtract should have been called for prerequisite")

    _ = app.InstallPluginModule("PHP", "Composer")
    assert.True(t, mockPlugin.InstallModuleCalled, "InstallModule should have been called")

    _ = app.UninstallPluginModule("PHP", "Composer")
    assert.True(t, mockPlugin.UninstallModuleCalled, "UninstallModule should have been called")

    _ = app.SwitchServiceVersion("PHP", "8.2.0")
    assert.True(t, mockPlugin.SwitchServiceVersionCalled, "SwitchServiceVersion should have been called")

    _ = app.DeleteVersion("PHP", "8.1.0")
    assert.True(t, mockPlugin.DeleteVersionCalled, "DeleteVersion should have been called")

    app.CancelDownload("test")
    assert.True(t, mockPlugin.CancelDownloadCalled, "CancelDownload should have been called")
}

func TestApp_SSHFunctions(t *testing.T) {
    app, _, _, _, mockSSH, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    _, _ = app.GetSSHSessions()
    assert.True(t, mockSSH.GetSessionsCalled, "GetSSHSessions should have been called")

    _ = app.AddSSHSession(config.SSHSession{ID: "test"})
    assert.True(t, mockSSH.AddSSHSessionCalled, "AddSSHSession should have been called")

    _ = app.UpdateSSHSession(config.SSHSession{ID: "test"})
    assert.True(t, mockSSH.UpdateSSHSessionCalled, "UpdateSSHSession should have been called")

    _ = app.DeleteSSHSession("test")
    assert.True(t, mockSSH.DeleteSSHSessionCalled, "DeleteSSHSession should have been called")

    _ = app.SaveSSHSessions([]config.SSHSession{})
    assert.True(t, mockSSH.SaveSessionsCalled, "SaveSSHSessions should have been called")

    _ = app.ConnectSSH(config.SSHSession{ID: "test"})
    assert.True(t, mockSSH.ConnectCalled, "ConnectSSH should have been called")

    app.DisconnectSSH("test")
    assert.True(t, mockSSH.DisconnectCalled, "DisconnectSSH should have been called")

    _ = app.SendSSHInput("test", "ls")
    assert.True(t, mockSSH.SendInputCalled, "SendSSHInput should have been called")

    _ = app.ResizeSSHTerminal("test", 80, 24)
    assert.True(t, mockSSH.ResizeTerminalCalled, "ResizeSSHTerminal should have been called")

    _, _ = app.GetRemoteFiles("test", "/")
    assert.True(t, mockSSH.ListFilesCalled, "GetRemoteFiles should have been called")

    _ = app.ExecuteSFTPAction("test", "mkdir", "/path", "")
    assert.True(t, mockSSH.ExecuteSFTPActionCalled, "ExecuteSFTPAction should have been called")

    _ = app.EditRemoteFile("test", "/path")
    assert.True(t, mockSSH.EditFileCalled, "EditRemoteFile should have been called")

    _, _ = app.GetRemoteCurrentPath("test")
    assert.True(t, mockSSH.GetCurrentPathCalled, "GetRemoteCurrentPath should have been called")

    _ = app.DownloadRemoteFile("test", "/path")
    assert.True(t, mockSSH.DownloadFileCalled, "DownloadRemoteFile should have been called")

    _ = app.UploadRemoteFile("test", "/path")
    assert.True(t, mockSSH.UploadFileCalled, "UploadRemoteFile should have been called")
}

func TestApp_ConfigFunctions(t *testing.T) {
    app, mockR, mockOrchestrator, _, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    // Test GetConfig
    cfg := app.GetConfig()
    assert.NotNil(t, cfg, "GetConfig should return a config")
    assert.Equal(t, tempDir, cfg.BaseDir, "BaseDir in config should match tempDir")

    // Test SelectDefaultEditor
    _, _ = app.SelectDefaultEditor()
    assert.True(t, mockR.OpenFileDialogCalled, "OpenFileDialog should be called for SelectDefaultEditor")
    mockR.OpenFileDialogCalled = false // Reset for next call

    // Test SetDefaultEditor
    editorPath := "/usr/bin/code"
    _ = app.SetDefaultEditor(editorPath)
    assert.Equal(t, editorPath, app.cfg.Editor, "Editor path should be set in config")

    // Test UpdateActiveTab
    app.UpdateActiveTab("activity")
    assert.True(t, mockOrchestrator.SetActiveTabCalled, "SetActiveTab should be called")

    // Test IsAdmin - Assuming it's an internal check, not delegating to a mock
    // If it were to delegate, we'd need a mock for it. For now, just call it.
    _ = app.IsAdmin()
}

func TestApp_EnvironmentFunctions(t *testing.T) {
    app, mockR, _, _, _, _, mockSymlink, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    // Test SetWWWRoot
    newWWWRoot := filepath.Join(tempDir, "new_www")
    _ = app.SetWWWRoot(newWWWRoot)
    assert.Equal(t, newWWWRoot, app.cfg.WWWRoot, "WWWRoot should be updated in config")

    // Test SetServerRoot
    newServerRoot := filepath.Join(tempDir, "new_server")
    _ = app.SetServerRoot(newServerRoot)
    assert.Equal(t, newServerRoot, app.cfg.BaseDir, "BaseDir (ServerRoot) should be updated in config")

    // Test SelectServerRoot
    _, _ = app.SelectServerRoot()
    assert.True(t, mockR.OpenDirectoryDialogCalled, "OpenDirectoryDialog should be called for SelectServerRoot")
    mockR.OpenDirectoryDialogCalled = false // Reset for next call

    // Test SelectWWWRoot
    _, _ = app.SelectWWWRoot()
    assert.True(t, mockR.OpenDirectoryDialogCalled, "OpenDirectoryDialog should be called for SelectWWWRoot")
    mockR.OpenDirectoryDialogCalled = false // Reset for next call

    // Test OpenServerRootFolder
    _ = app.OpenServerRootFolder()
    assert.True(t, mockSymlink.OpenFolderCalled, "OpenFolder should be called for OpenServerRootFolder")
    mockSymlink.OpenFolderCalled = false // Reset for next call

    // Test OpenAppsLocationFolder
    _ = app.OpenAppsLocationFolder()
    assert.True(t, mockSymlink.OpenFolderCalled, "OpenFolder should be called for OpenAppsLocationFolder")
}

func TestApp_NetworkFunctions(t *testing.T) {
    app, _, _, _, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    // Test CheckProxyPorts - Assuming it's an internal check, not delegating to a mock
    // If it were to delegate, we'd need a mock for it. For now, just call it.
    _ = app.CheckProxyPorts()

    // Test GetProxyApps
    proxies := app.GetProxyApps()
    assert.Contains(t, proxies, "test", "GetProxyApps should return existing proxies")
    assert.Equal(t, 3000, proxies["test"], "Proxy port should match initial config")

    // Test SaveProxyPort
    _ = app.SaveProxyPort("newapp", 5000)
    assert.Equal(t, 5000, app.cfg.Proxies["newapp"], "New proxy port should be saved")

    // Test OpenProxyTerminal - Assuming it's an internal call, not delegating to a mock
    // If it were to delegate, we'd need a mock for it. For now, just call it.
    app.OpenProxyTerminal("myapp", "cmd")
}

func TestApp_PHPFunctions(t *testing.T) {
    app, _, _, _, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    // These functions are not in the provided app.go snippet,
    // so we can only call them if they exist and don't panic.
    // If they delegate to a mock, we'd add assertions for that mock.
    // For now, just calling them.
    _, _ = app.GetPHPExtensions()
    _ = app.TogglePHPExtension("openssl", true)
}

func TestApp_ProfileFunctions(t *testing.T) {
    app, _, _, _, _, _, _, tempDir := setupApp(t)
    defer teardownApp(tempDir)

    // These functions are complex and involve file I/O.
    // For now, just call them to ensure they don't panic.
    // Proper testing would require mocking file system operations or
    // verifying specific file content changes.
    _ = app.ExportProfile(true, true)
    _ = app.ImportProfile()
}

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil { t.Error("NewApp returned nil") }
}
