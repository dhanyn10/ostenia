package backend

import (
	"context"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/service"
	"testing"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"os/exec"
    "os"
    "path/filepath"
)

type MockRuntime struct {
}

func (m *MockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {}
func (m *MockRuntime) WindowMinimise(ctx context.Context)   {}
func (m *MockRuntime) WindowMaximise(ctx context.Context)   {}
func (m *MockRuntime) WindowUnmaximise(ctx context.Context) {}
func (m *MockRuntime) WindowExecJS(ctx context.Context, js string) {}
func (m *MockRuntime) Quit(ctx context.Context)             {}
func (m *MockRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return "C:\\selected\\file.txt", nil
}
func (m *MockRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return "C:\\selected", nil
}
func (m *MockRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return "C:\\selected\\saved.txt", nil
}

type MockExecutor struct {
}
func (m *MockExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command("echo", "success")
}

func TestApp_Full(t *testing.T) {
	oldExecutor := utils.Executor
	utils.Executor = &MockExecutor{}
	defer func() { utils.Executor = oldExecutor }()

	mock := &MockRuntime{}
    tempDir, _ := os.MkdirTemp("", "ostenia-app-*")
    defer os.RemoveAll(tempDir)
    os.Setenv("OSTENIA_HOME", tempDir)

	app := &App{
		runtime:      mock,
		ctx:          context.Background(),
		cfg:          &config.Config{Proxies: map[string]int{"test": 3000}, BaseDir: tempDir, WWWRoot: filepath.Join(tempDir, "www")},
		downloader:   plugins.NewManager(nil),
		orchestrator: service.NewOrchestrator(nil),
		sshManager:   service.NewSSHManager(nil),
	}
    app.orchestrator.SetRuntime(mock)

    // Setup dummy bin files
    services := []string{"apache", "mysql", "nginx", "php", "nodejs", "python"}
    exes := []string{"httpd.exe", "mysqld.exe", "nginx.exe", "php-cgi.exe", "node.exe", "python.exe"}
    for i, s := range services {
        d := filepath.Join(tempDir, "bin", s, "current")
        if s == "nginx" {
             os.MkdirAll(d, 0755)
             os.WriteFile(filepath.Join(d, exes[i]), []byte(""), 0644)
        } else {
             os.MkdirAll(filepath.Join(d, "bin"), 0755)
             os.WriteFile(filepath.Join(d, "bin", exes[i]), []byte(""), 0644)
        }
    }
    // PHP specific
    os.WriteFile(filepath.Join(tempDir, "bin", "php", "current", "php.exe"), []byte(""), 0644)
    os.WriteFile(filepath.Join(tempDir, "bin", "php", "current", "php.ini"), []byte(""), 0644)

	app.Minimize()
	app.Maximize()
	app.Unmaximize()
	app.ToggleDevTools()
	app.Close()

	_ = app.GetConfig()
	_, _ = app.SelectDefaultEditor()
	_ = app.SetDefaultEditor("path")
	app.UpdateActiveTab("activity")
	_ = app.IsAdmin()

	_ = app.SetWWWRoot(filepath.Join(tempDir, "www2"))
	_ = app.SetServerRoot(tempDir)
	_, _ = app.SelectServerRoot()
	_, _ = app.SelectWWWRoot()
	_ = app.OpenServerRootFolder()
	_ = app.OpenAppsLocationFolder()
	app.OpenTerminal("cmd")

	_ = app.GetPrerequisites()
	app.CancelDownload("test")
	_ = app.OpenPluginFolder("PHP")
    _ = app.InstallPluginModule("PHP", "Composer")
    _ = app.UninstallPluginModule("PHP", "Composer")
    _ = app.SwitchServiceVersion("PHP", "8.2.0")
    _ = app.DeleteVersion("PHP", "8.1.0")

	_ = app.GetServiceStatus("Apache")
	_ = app.StartService("Apache")
    _ = app.StartService("MySQL")
    _ = app.StartService("Nginx")
    _ = app.StartService("PHP")
    _ = app.StartService("Node.js")
    _ = app.StartService("Python")
    _ = app.StartService("OpenSSL")
	_ = app.StopService("Apache")
    _ = app.StopService("OpenSSL")
    _ = app.StopService("PHP")
    _ = app.StopService("Node.js")
    _ = app.StopService("Python")
	_ = app.StartAllServices()
	app.StopAllServices()
	_ = app.SetApacheHTTPS(true)
	_ = app.SetNginxHTTPS(true)
    _ = app.OpenHeidiSQL()
    _ = app.OpenServiceTerminal("Apache", "cmd")

	_, _ = app.GetSSHSessions()
	_ = app.AddSSHSession(config.SSHSession{ID: "test"})
	_ = app.UpdateSSHSession(config.SSHSession{ID: "test"})
	_ = app.DeleteSSHSession("test")
	_ = app.SaveSSHSessions([]config.SSHSession{})
	_ = app.DownloadRemoteFile("test", "/path")
	_ = app.UploadRemoteFile("test", "/path")
    _ = app.ConnectSSH(config.SSHSession{ID: "test"})
    app.DisconnectSSH("test")
    _ = app.SendSSHInput("test", "ls")
    _ = app.ResizeSSHTerminal("test", 80, 24)
    _, _ = app.GetRemoteFiles("test", "/")
    _ = app.ExecuteSFTPAction("test", "mkdir", "/path", "")
    _ = app.EditRemoteFile("test", "/path")
    _, _ = app.GetRemoteCurrentPath("test")

    // app_network.go
    _ = app.CheckProxyPorts()
    _ = app.GetProxyApps()
    _ = app.SaveProxyPort("myapp", 4000)

    // app_php.go
    _, _ = app.GetPHPExtensions()
    _ = app.TogglePHPExtension("openssl", true)

    // app_profile.go
    _ = app.ExportProfile(true, true)
    _ = app.ImportProfile()
}

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil { t.Error("NewApp returned nil") }
    // Test Startup with dummy environment
    tempDir, _ := os.MkdirTemp("", "ostenia-startup-*")
    defer os.RemoveAll(tempDir)
    os.Setenv("OSTENIA_HOME", tempDir)
    a.Startup(context.Background())
}
