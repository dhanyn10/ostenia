package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"ostenia/internal/backend"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	plugins_utils "ostenia/internal/plugins/utils"
	"ostenia/internal/service"
	"path/filepath"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct manages the main application state and coordinates between backend services and the frontend
type App struct {
	ctx            context.Context
	downloader     *plugins.Manager
	orchestrator   *service.Orchestrator
	symlinkMgr     *service.SymlinkManager
	sshManager     *service.SSHManager
	cfg            *config.Config
	serviceMgr     *backend.ServiceManager
	terminalMgr    *backend.TerminalManager
	sshMgr         *backend.SSHManagerDelegate
	configMgr      *backend.ConfigManager
	profileMgr     *backend.ProfileManager
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.downloader = plugins.NewManager(ctx)
	a.orchestrator = service.NewOrchestrator(ctx)
	a.symlinkMgr = service.NewSymlinkManager()
	a.sshManager = service.NewSSHManager(ctx)

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("[App] Error loading config: %v\n", err)
	}
	a.cfg = cfg

	a.serviceMgr = &backend.ServiceManager{
		Ctx:          ctx,
		Downloader:   a.downloader,
		Orchestrator: a.orchestrator,
		SymlinkMgr:   a.symlinkMgr,
		Cfg:          a.cfg,
	}

	a.terminalMgr = &backend.TerminalManager{
		Ctx:            ctx,
		ServiceManager: a.serviceMgr,
		Cfg:            a.cfg,
	}

	a.sshMgr = &backend.SSHManagerDelegate{
		SSHManager: a.sshManager,
		Config:     a.cfg,
	}

	a.configMgr = &backend.ConfigManager{
		Ctx:    ctx,
		Config: a.cfg,
	}

	a.profileMgr = &backend.ProfileManager{
		Ctx:    ctx,
		Config: a.cfg,
	}

	// Initial setup of directories in current base dir
	a.ensureEnvironmentStructure()

	// Start the periodic watcher for services
	a.orchestrator.StartWatcher()

	// Start proxy port watcher
	go a.startProxyWatcher()
}

// ToggleDevTools toggles the browser developer tools
func (a *App) ToggleDevTools() {
	wruntime.WindowExecJS(a.ctx, "window.runtime.WindowToggleDevTools()")
}

// Minimize minimizes the application window
func (a *App) Minimize() { wruntime.WindowMinimise(a.ctx) }

// Maximize maximizes the application window
func (a *App) Maximize() { wruntime.WindowMaximise(a.ctx) }

// Unmaximize restores the application window from maximized state
func (a *App) Unmaximize() { wruntime.WindowUnmaximise(a.ctx) }

// Close closes the application
func (a *App) Close() { wruntime.Quit(a.ctx) }

func (a *App) startProxyWatcher() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			statuses := a.CheckProxyPorts()
			wruntime.EventsEmit(a.ctx, "proxy_status", statuses)
		}
	}
}

func (a *App) ensureEnvironmentStructure() {
	baseDir := config.GetBaseDir()
	dirs := []string{filepath.Join(baseDir, "bin"), filepath.Join(baseDir, "ssl")}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_ = os.MkdirAll(dir, 0755)
		}
	}
	if a.cfg != nil && a.cfg.WWWRoot != "" {
		_ = os.MkdirAll(a.cfg.WWWRoot, 0755)
	}
}

// IsAdmin checks if the application is running with administrative privileges
func (a *App) IsAdmin() bool { return service.IsAdmin() }

// GetInstalledApps returns the list of installed editors/apps for external editing
func (a *App) GetInstalledApps() ([]backend.InstalledApp, error) {
	return backend.GetInstalledApps()
}

// GetPrerequisites returns the list of latest known plugin versions for installation
func (a *App) GetPrerequisites() []plugins.DownloadTask { return plugins.GetLatestKnownVersions() }

// GetServiceStatus returns detailed information about a specific service
func (a *App) GetServiceStatus(serviceName string) service.ServiceDetailedInfo {
	return a.orchestrator.GetDetailedInfo(serviceName)
}

// OpenHeidiSQL launches the HeidiSQL application if installed
func (a *App) OpenHeidiSQL() error {
	exePath, _ := plugins.DetectHeidiSQLInstallation()
	if exePath == "" {
		return fmt.Errorf("HeidiSQL is not installed")
	}
	cmdPath := filepath.Join(plugins_utils.GetSystemDirectory(), "cmd.exe")
	cmd := exec.Command(cmdPath, "/c", "start", "", exePath)
	cmd.Env = plugins_utils.SafeEnv()
	plugins_utils.SetHideWindow(cmd)
	return cmd.Run()
}

// SetWWWRoot sets the server root directory (www)
func (a *App) SetWWWRoot(path string) error {
	a.cfg.WWWRoot = path
	err := config.SaveConfig(a.cfg)
	if err != nil { return err }
	_ = os.MkdirAll(path, 0755)
	if a.orchestrator.IsRunning("Apache") {
		_ = a.StopService("Apache")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.StopService("Nginx")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Nginx")
	}
	return nil
}

// ProxyStatusInfo represents the health status of a proxy target
type ProxyStatusInfo struct {
	Name   string `json:"name"`
	IsUp   bool   `json:"isUp"`
	Port   int    `json:"port"`
}

// CheckProxyPorts checks if the configured proxy ports are reachable
func (a *App) CheckProxyPorts() []ProxyStatusInfo {
	var statuses []ProxyStatusInfo
	for name, port := range a.cfg.Proxies {
		isUp := false
		if port > 0 {
			timeout := 500 * time.Millisecond
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
			if err == nil {
				isUp = true
				conn.Close()
			}
		}
		statuses = append(statuses, ProxyStatusInfo{Name: name, IsUp: isUp, Port: port})
	}
	return statuses
}

// OpenProxyTerminal opens a terminal at the directory of a proxy app
func (a *App) OpenProxyTerminal(name string, terminalType string) error {
	return a.terminalMgr.OpenProxyTerminal(name, terminalType)
}

// SetServerRoot changes the base directory for all Ostenia apps and binaries
func (a *App) SetServerRoot(rootPath string) error {
	a.orchestrator.StopAll()
	time.Sleep(1 * time.Second)
	a.cfg.BaseDir = rootPath
	a.cfg.WWWRoot = filepath.Join(rootPath, "www")
	err := config.SaveConfig(a.cfg)
	if err != nil { return err }
	a.ensureEnvironmentStructure()
	a.orchestrator.RequestRefresh()
	wruntime.EventsEmit(a.ctx, "environment_changed", a.cfg)
	return nil
}

// SelectServerRoot opens a directory dialog to select the Ostenia apps location
func (a *App) SelectServerRoot() (string, error) {
	selectedDir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Ostenia Apps Location"})
	if err != nil { return "", err }
	if selectedDir != "" { err = a.SetServerRoot(selectedDir); if err != nil { return "", err } }
	return selectedDir, nil
}

// SelectWWWRoot opens a directory dialog to select the server root (www)
func (a *App) SelectWWWRoot() (string, error) {
	selectedDir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Server Root (www)"})
	if err != nil { return "", err }
	if selectedDir != "" { err = a.SetWWWRoot(selectedDir); if err != nil { return "", err } }
	return selectedDir, nil
}

// OpenServerRootFolder opens the www directory in File Explorer
func (a *App) OpenServerRootFolder() error { return service.OpenExplorer(a.cfg.WWWRoot) }

// OpenAppsLocationFolder opens the Ostenia base directory in File Explorer
func (a *App) OpenAppsLocationFolder() error { return service.OpenExplorer(config.GetBaseDir()) }

// OpenPluginFolder opens the binary directory for a specific service in File Explorer
func (a *App) OpenPluginFolder(serviceName string) error {
	_, binDir, _ := a.serviceMgr.GetPluginPaths(serviceName)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		_ = os.MkdirAll(binDir, 0755)
	}
	return service.OpenExplorer(binDir)
}

// InstallPrerequisite downloads and installs a plugin prerequisite
func (a *App) InstallPrerequisite(task plugins.DownloadTask) error {
	return a.serviceMgr.InstallPrerequisite(task)
}

// CancelDownload cancels an ongoing plugin download
func (a *App) CancelDownload(taskName string) { a.downloader.CancelDownload(taskName) }

// InstallPluginModule installs a sub-module for a parent plugin (e.g., Composer for PHP)
func (a *App) InstallPluginModule(parentName string, moduleName string) error {
	return a.serviceMgr.InstallPluginModule(parentName, moduleName)
}

// UninstallPluginModule removes a sub-module from a parent plugin
func (a *App) UninstallPluginModule(parentName string, moduleName string) error {
	return a.serviceMgr.UninstallPluginModule(parentName, moduleName)
}

// StartService starts a background service by name
func (a *App) StartService(serviceName string) error {
	return a.serviceMgr.StartService(serviceName)
}

// StopService stops a running service by name
func (a *App) StopService(serviceName string) error {
	return a.serviceMgr.StopService(serviceName)
}

// SwitchServiceVersion changes the active version of a service using directory junctions
func (a *App) SwitchServiceVersion(serviceName string, version string) error {
	return a.serviceMgr.SwitchServiceVersion(serviceName, version)
}

// StartAllServices starts the default stack (MySQL, PHP, Apache)
func (a *App) StartAllServices() error {
	_ = a.StartService("MySQL")
	_ = a.StartService("PHP")
	return a.StartService("Apache")
}

// StopAllServices stops all currently running background services
func (a *App) StopAllServices() { a.orchestrator.StopAll() }

// OpenTerminal opens a terminal at the current server root directory
func (a *App) OpenTerminal(terminalType string) {
	a.terminalMgr.OpenTerminal(terminalType)
}

// OpenTerminalAtPath opens a terminal at a specific local path with the Ostenia environment variables set
func (a *App) OpenTerminalAtPath(terminalType string, path string) {
	a.terminalMgr.OpenTerminalAtPath(terminalType, path)
}

// DeleteVersion deletes a specific version folder of a plugin
func (a *App) DeleteVersion(serviceName string, version string) error {
	return a.downloader.DeleteVersion(serviceName, version)
}

// SetApacheHTTPS enables or disables HTTPS support for Apache
func (a *App) SetApacheHTTPS(enabled bool) error {
	return a.serviceMgr.SetApacheHTTPS(enabled)
}

// SetNginxHTTPS enables or disables HTTPS support for Nginx
func (a *App) SetNginxHTTPS(enabled bool) error {
	return a.serviceMgr.SetNginxHTTPS(enabled)
}

// OpenServiceTerminal opens a terminal at the binary directory of a specific service
func (a *App) OpenServiceTerminal(serviceName string, terminalType string) error {
	return a.terminalMgr.OpenServiceTerminal(serviceName, terminalType)
}

// GetPHPExtensions returns the list of PHP extensions and their status from php.ini
func (a *App) GetPHPExtensions() ([]service.PHPExtensionInfo, error) {
	baseDir := config.GetBaseDir(); phpPath := filepath.Join(baseDir, "bin", "php", "current"); return service.GetPHPExtensions(phpPath)
}

// TogglePHPExtension enables or disables a PHP extension in php.ini
func (a *App) TogglePHPExtension(extName string, enable bool) error {
	baseDir := config.GetBaseDir(); phpPath := filepath.Join(baseDir, "bin", "php", "current"); err := service.TogglePHPExtension(phpPath, extName, enable); if err != nil { return err }
	if a.orchestrator.IsRunning("PHP") {
		_ = a.StopService("PHP")
		time.Sleep(600 * time.Millisecond)
		return a.StartService("PHP")
	}
	return nil
}

// ProxyAppInfo represents basic information about a potential proxy app directory
type ProxyAppInfo struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// GetProxyApps scans the www directory and returns a list of folders and their configured proxy ports
func (a *App) GetProxyApps() []ProxyAppInfo {
	var apps []ProxyAppInfo
	entries, err := os.ReadDir(a.cfg.WWWRoot)
	if err != nil { return apps }
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			port := 0
			if p, ok := a.cfg.Proxies[name]; ok { port = p }
			apps = append(apps, ProxyAppInfo{Name: name, Port: port})
		}
	}
	return apps
}

// GetConfig returns the current application configuration
func (a *App) GetConfig() (*config.Config, error) { return a.configMgr.GetConfig() }

// SetDefaultEditor sets the path to the default external text editor
func (a *App) SetDefaultEditor(editorPath string) error { return a.configMgr.SetDefaultEditor(editorPath) }

// SelectDefaultEditor opens a file dialog to select the default text editor
func (a *App) SelectDefaultEditor() (string, error) { return a.configMgr.SelectDefaultEditor() }

// ExportProfile exports the application configuration and/or SSH sessions to a JSON file
func (a *App) ExportProfile(includeConfig bool, includeSSH bool) error {
	return a.profileMgr.ExportProfile(includeConfig, includeSSH)
}

// ImportProfile imports an Ostenia profile from a JSON file
func (a *App) ImportProfile() error { return a.profileMgr.ImportProfile() }

// GetSSHSessions returns the list of saved SSH sessions
func (a *App) GetSSHSessions() ([]config.SSHSession, error) { return config.LoadSSHSessions() }

// SaveSSHSessions saves the entire list of SSH sessions
func (a *App) SaveSSHSessions(sessions []config.SSHSession) error { return config.SaveSSHSessions(sessions) }

// AddSSHSession adds a new SSH session
func (a *App) AddSSHSession(session config.SSHSession) error { return config.AddSSHSession(session) }

// UpdateSSHSession updates an existing SSH session
func (a *App) UpdateSSHSession(session config.SSHSession) error { return config.UpdateSSHSession(session) }

// DeleteSSHSession removes an SSH session by ID
func (a *App) DeleteSSHSession(id string) error { return config.DeleteSSHSession(id) }

// ConnectSSH initiates an SSH connection
func (a *App) ConnectSSH(session config.SSHSession) error { return a.sshMgr.Connect(session) }

// DisconnectSSH closes an SSH connection
func (a *App) DisconnectSSH(sessionID string) { a.sshMgr.Disconnect(sessionID) }

// SendSSHInput sends terminal input to an active SSH session
func (a *App) SendSSHInput(sessionID string, data string) error { return a.sshMgr.SendInput(sessionID, data) }

// ResizeSSHTerminal updates the PTY size for an SSH session
func (a *App) ResizeSSHTerminal(sessionID string, cols int, rows int) error {
	return a.sshMgr.ResizeTerminal(sessionID, cols, rows)
}

// GetRemoteFiles lists files in a remote directory via SFTP
func (a *App) GetRemoteFiles(sessionID string, path string) ([]service.RemoteFile, error) {
	return a.sshMgr.ListFiles(sessionID, path)
}

// ExecuteSFTPAction performs file operations (rename, delete, mkdir) on a remote host
func (a *App) ExecuteSFTPAction(sessionID string, action string, path string, target string) error {
	return a.sshMgr.ExecuteSFTPAction(sessionID, action, path, target)
}

// EditRemoteFile downloads a remote file to a temporary location and opens it in the default editor
func (a *App) EditRemoteFile(sessionID string, remotePath string) error {
	return a.sshMgr.EditFile(sessionID, remotePath)
}

// GetRemoteCurrentPath returns the current working directory of an SSH session
func (a *App) GetRemoteCurrentPath(sessionID string) (string, error) {
	return a.sshMgr.GetCurrentPath(sessionID)
}

// DownloadRemoteFile downloads a file from a remote host to the local machine
func (a *App) DownloadRemoteFile(sessionID string, remotePath string) error {
	fileName := filepath.Base(remotePath)
	localPath, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Download File",
		DefaultFilename: fileName,
	})
	if err != nil || localPath == "" { return err }
	return a.sshMgr.DownloadFile(sessionID, remotePath, localPath)
}

// UploadRemoteFile uploads a file from the local machine to a remote host
func (a *App) UploadRemoteFile(sessionID string, remoteDir string) error {
	localPath, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Upload File"})
	if err != nil || localPath == "" { return err }
	remotePath := filepath.ToSlash(filepath.Join(remoteDir, filepath.Base(localPath)))
	return a.sshMgr.UploadFile(sessionID, localPath, remotePath)
}

// SaveProxyPort saves the proxy port for a specific folder and reconfigures web servers
func (a *App) SaveProxyPort(name string, port int) error {
	if a.cfg.Proxies == nil { a.cfg.Proxies = make(map[string]int) }
	if port <= 0 { delete(a.cfg.Proxies, name) } else { a.cfg.Proxies[name] = port }
	err := config.SaveConfig(a.cfg)
	if err != nil { return err }

	if a.orchestrator.IsRunning("Apache") {
		_ = a.serviceMgr.UpdateApacheConfig(filepath.Join(config.GetBaseDir(), "bin", "apache", "current"), a.orchestrator.GetDetailedInfo("Apache").Port)
		_ = a.StopService("Apache")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.serviceMgr.UpdateNginxConfig(filepath.Join(config.GetBaseDir(), "bin", "nginx", "current"), a.orchestrator.GetDetailedInfo("Nginx").Port)
		_ = a.StopService("Nginx")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Nginx")
	}
	return nil
}
