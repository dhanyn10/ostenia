package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	plugins_utils "ostenia/internal/plugins/utils"
	"ostenia/internal/plugins/php"
	"ostenia/internal/plugins/python"
	"ostenia/internal/network"
	"ostenia/internal/service"
	"ostenia/internal/ssl"
	"path/filepath"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct manages the main application state and coordinates between backend services and the frontend
type App struct {
	ctx          context.Context
	downloader   *plugins.Manager
	orchestrator *service.Orchestrator
	symlinkMgr   *service.SymlinkManager
	sshManager   *service.SSHManager
	cfg          *config.Config
}

const (
	exeNginx  = "nginx.exe"
	exeApache = "httpd.exe"
	exeMySQL  = "mysqld.exe"
	exePHP    = "php-cgi.exe"
	exeNode   = "node.exe"
	exePython = "python.exe"
)

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
		// Log error if config fails to load
		fmt.Printf("[App] Error loading config: %v\n", err)
	}
	a.cfg = cfg

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
	fmt.Printf("[App] Setting Server Root (www) to: %s\n", path)
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
	path := filepath.Join(a.cfg.WWWRoot, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("folder %s not found", name)
	}
	a.OpenTerminalAtPath(terminalType, path)
	return nil
}

// SetServerRoot changes the base directory for all Ostenia apps and binaries
func (a *App) SetServerRoot(rootPath string) error {
	fmt.Printf("[App] Switching Apps Location to: %s\n", rootPath)
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

func (a *App) getPluginPaths(serviceName string) (category string, binDir string, currentPath string) {
	category = strings.ToLower(serviceName)
	if category == "node.js" {
		category = "nodejs"
	}
	binDir = filepath.Join(config.GetBaseDir(), "bin", category)
	currentPath = filepath.Join(binDir, "current")
	return
}

// OpenPluginFolder opens the binary directory for a specific service in File Explorer
func (a *App) OpenPluginFolder(serviceName string) error {
	_, binDir, _ := a.getPluginPaths(serviceName)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		_ = os.MkdirAll(binDir, 0755)
	}
	return service.OpenExplorer(binDir)
}

// InstallPrerequisite downloads and installs a plugin prerequisite
func (a *App) InstallPrerequisite(task plugins.DownloadTask) error {
	err := a.downloader.DownloadAndExtract(task)
	if err == nil {
		_, _, currentPath := a.getPluginPaths(task.Name)

		if task.Name == "PHP" {
			_ = service.UpdatePHPPath(currentPath, true)
		} else if task.Name == "Python" {
			_ = service.UpdatePythonPath(currentPath, true)
		}
		a.orchestrator.RequestRefresh()
	}
	return err
}

// CancelDownload cancels an ongoing plugin download
func (a *App) CancelDownload(taskName string) { a.downloader.CancelDownload(taskName) }

// InstallPluginModule installs a sub-module for a parent plugin (e.g., Composer for PHP)
func (a *App) InstallPluginModule(parentName string, moduleName string) error {
	_, _, currentPath := a.getPluginPaths(parentName)

	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("%s is not installed or active", parentName)
	}

	emitProgress := func(name string, pct float64, status string) {
		wruntime.EventsEmit(a.ctx, "download_progress", plugins.Progress{Name: name, Percentage: pct, Status: status})
	}

	var err error
	switch parentName {
	case "PHP":
		err = php.InstallModule(a.ctx, a.downloader, moduleName, currentPath, emitProgress)
		if err == nil {
			_ = service.UpdatePHPPath(currentPath, true)
		}
	case "Python":
		err = python.InstallModule(a.ctx, a.downloader, moduleName, currentPath, emitProgress)
		if err == nil {
			_ = service.UpdatePythonPath(currentPath, true)
		}
	default:
		err = fmt.Errorf("unsupported parent plugin: %s", parentName)
	}

	if err == nil {
		a.orchestrator.RequestRefresh()
	}
	return err
}

// UninstallPluginModule removes a sub-module from a parent plugin
func (a *App) UninstallPluginModule(parentName string, moduleName string) error {
	_, _, currentPath := a.getPluginPaths(parentName)

	var err error
	switch parentName {
	case "PHP":
		err = php.UninstallModule(moduleName, currentPath)
		if err == nil {
			_ = service.UpdatePHPPath(currentPath, true)
		}
	case "Python":
		err = python.UninstallModule(moduleName, currentPath)
		if err == nil {
			_ = service.UpdatePythonPath(currentPath, true)
		}
	default:
		err = fmt.Errorf("unsupported parent plugin: %s", parentName)
	}

	if err == nil {
		a.orchestrator.RequestRefresh()
	}
	return err
}

// StartService starts a background service by name
func (a *App) StartService(serviceName string) error {
	_, binDir, currentPath := a.getPluginPaths(serviceName)
	fmt.Printf("[App] Starting service: %s\n", serviceName)

	switch serviceName {
	case "Node.js":
		return a.startNodeService(currentPath)
	case "Python":
		return a.startPythonService(currentPath)
	case "OpenSSL":
		return a.startOpenSSLService()
	case "MySQL":
		return a.startMySQLService(binDir)
	case "Apache":
		return a.startApacheService(binDir)
	case "Nginx":
		return a.startNginxService(binDir)
	case "HeidiSQL":
		return a.startHeidiSQLService()
	case "PHP":
		return a.startPHPService(currentPath)
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

func (a *App) startNodeService(currentPath string) error {
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("node.js not installed")
	}
	if err := service.UpdateNodePath(currentPath, true); err != nil {
		return err
	}
	a.orchestrator.RequestRefresh()
	return nil
}

func (a *App) startPythonService(currentPath string) error {
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("python not installed")
	}
	if err := service.UpdatePythonPath(currentPath, true); err != nil {
		return err
	}
	a.orchestrator.RequestRefresh()
	return nil
}

func (a *App) startOpenSSLService() error {
	caDir := filepath.Join(config.GetBaseDir(), "ssl")
	if err := ssl.GenerateRootCA(caDir); err != nil {
		return err
	}
	a.orchestrator.RequestRefresh()
	return nil
}

func (a *App) findExecutable(binDir string, exeName string) (string, string) {
	currentPath := filepath.Join(binDir, "current")

	// 1. Try "current" link first for efficiency
	if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		path := filepath.Join(resolved, "bin", exeName)
		if exeName == exeNginx {
			path = filepath.Join(resolved, exeName)
		}
		if _, err := os.Stat(path); err == nil {
			return path, resolved
		}
		// Apache fallback
		if exeName == exeApache {
			path = filepath.Join(resolved, "Apache24", "bin", exeName)
			if _, err := os.Stat(path); err == nil {
				return path, filepath.Join(resolved, "Apache24")
			}
		}
	}

	// 2. Fallback to Walk if "current" is not valid or doesn't match
	var binPath, basePath string
	_ = filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == exeName {
			binPath = path
			basePath = filepath.Dir(filepath.Dir(path))
			if exeName == exeNginx {
				basePath = filepath.Dir(path)
			}
			return filepath.SkipDir
		}
		return nil
	})
	return binPath, basePath
}

func (a *App) startMySQLService(binDir string) error {
	mysqlBin, mysqlBase := a.findExecutable(binDir, exeMySQL)
	if mysqlBin == "" {
		return fmt.Errorf("%s not found", exeMySQL)
	}
	port := a.orchestrator.GetDetailedInfo("MySQL").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{3306, 3307, 3308})
		if err != nil {
			return err
		}
		port = p
	}
	_ = os.Setenv("PATH", filepath.Dir(mysqlBin)+";"+os.Getenv("PATH"))
	if err := a.updateMySQLConfig(mysqlBase, port); err != nil {
		return err
	}
	iniPath := filepath.Join(mysqlBase, "my.ini")
	return a.orchestrator.StartServiceWithPort("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath, "--console"}, filepath.Dir(mysqlBin), port)
}

func (a *App) startApacheService(binDir string) error {
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.StopService("Nginx")
		time.Sleep(600 * time.Millisecond)
	}
	apacheBin, apacheBase := a.findExecutable(binDir, exeApache)
	if apacheBase == "" {
		return fmt.Errorf("apache installation not found")
	}
	port := a.orchestrator.GetDetailedInfo("Apache").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil {
			return err
		}
		port = p
	}
	_, _, phpPath := a.getPluginPaths("PHP")
	_, _, nodePath := a.getPluginPaths("Node.js")
	_ = os.Setenv("PATH", phpPath+";"+os.Getenv("PATH")+";"+nodePath)
	if err := a.updateApacheConfig(apacheBase, port); err != nil {
		return err
	}
	return a.orchestrator.StartServiceWithPort("Apache", apacheBin, []string{}, apacheBase, port)
}

func (a *App) startNginxService(binDir string) error {
	if a.orchestrator.IsRunning("Apache") {
		_ = a.StopService("Apache")
		time.Sleep(600 * time.Millisecond)
	}
	nginxBin, nginxBase := a.findExecutable(binDir, exeNginx)
	if nginxBase == "" {
		return fmt.Errorf("nginx installation not found")
	}
	port := a.orchestrator.GetDetailedInfo("Nginx").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil {
			return err
		}
		port = p
	}
	if err := a.updateNginxConfig(nginxBase, port); err != nil {
		return err
	}
	return a.orchestrator.StartServiceWithPort("Nginx", nginxBin, []string{"-p", nginxBase}, nginxBase, port)
}

func (a *App) startHeidiSQLService() error {
	exePath, _ := plugins.DetectHeidiSQLInstallation()
	if exePath != "" {
		a.orchestrator.RequestRefresh()
		return nil
	}
	tasks := plugins.GetLatestKnownVersions()
	for _, t := range tasks {
		if t.Name == "HeidiSQL" {
			return a.downloader.DownloadAndExtract(t)
		}
	}
	return fmt.Errorf("HeidiSQL task not found")
}

func (a *App) startPHPService(currentPath string) error {
	phpCgi := filepath.Join(currentPath, exePHP)
	if _, err := os.Stat(phpCgi); os.IsNotExist(err) {
		return fmt.Errorf("%s not found", exePHP)
	}
	_ = service.UpdatePHPConfig(currentPath)
	port := a.orchestrator.GetDetailedInfo("PHP").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{9000, 9001, 9002})
		if err != nil {
			return err
		}
		port = p
	}
	_ = os.Setenv("PATH", currentPath+";"+os.Getenv("PATH"))
	_ = service.UpdatePHPPath(currentPath, true)
	_ = os.Setenv("PHP_FCGI_MAX_REQUESTS", "1000")
	err := a.orchestrator.StartServiceWithPort("PHP", phpCgi, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, currentPath, port)
	if err == nil {
		a.restartDependentWebServers()
	}
	return err
}

func (a *App) restartDependentWebServers() {
	for _, srv := range []string{"Apache", "Nginx"} {
		if a.orchestrator.IsRunning(srv) {
			_ = a.StopService(srv)
			time.Sleep(600 * time.Millisecond)
			_ = a.StartService(srv)
		}
	}
}

// StopService stops a running service by name
func (a *App) StopService(serviceName string) error {
	_, _, currentPath := a.getPluginPaths(serviceName)

	if serviceName == "OpenSSL" {
		baseDir := config.GetBaseDir()
		caDir := filepath.Join(baseDir, "ssl")
		_ = os.RemoveAll(caDir)
		_ = os.MkdirAll(caDir, 0755)
		_ = a.SetApacheHTTPS(false)
		_ = a.SetNginxHTTPS(false)
		a.orchestrator.RequestRefresh()
		return nil
	}
	if serviceName == "PHP" {
		_ = service.UpdatePHPPath(currentPath, false)
	}
	if serviceName == "Node.js" {
		_ = service.UpdateNodePath(currentPath, false)
		a.orchestrator.RequestRefresh()
		return nil
	}
	if serviceName == "Python" {
		_ = service.UpdatePythonPath(currentPath, false)
		a.orchestrator.RequestRefresh()
		return nil
	}
	return a.orchestrator.StopService(serviceName)
}

// SwitchServiceVersion changes the active version of a service using directory junctions
func (a *App) SwitchServiceVersion(serviceName string, version string) error {
	category, binDir, currentPath := a.getPluginPaths(serviceName)
	prefix := ""
	switch category {
	case "php":
		prefix = "php-"
	case "apache":
		prefix = "httpd-"
	case "mysql":
		prefix = "mysql-"
	case "nginx":
		prefix = "nginx-"
	case "nodejs":
		prefix = "node-v"
	case "python":
		prefix = "python-"
	}
	targetDir := filepath.Join(binDir, prefix+version)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		targetDir = filepath.Join(binDir, version)
	}
	wasRunning := a.orchestrator.IsRunning(serviceName)
	if wasRunning {
		_ = a.StopService(serviceName)
		time.Sleep(600 * time.Millisecond)
	}
	_ = os.Remove(currentPath)
	if _, err := os.Stat(targetDir); err == nil {
		cmdPath := filepath.Join(plugins_utils.GetSystemDirectory(), "cmd.exe")
		cmd := exec.Command(cmdPath, "/c", "mklink", "/J", currentPath, targetDir)
		cmd.Env = plugins_utils.SafeEnv()
		plugins_utils.SetHideWindow(cmd)
		_ = cmd.Run()
	}
	if category == "php" {
		_ = os.Setenv("PATH", currentPath+";"+os.Getenv("PATH"))
		_ = service.UpdatePHPConfig(currentPath)
		_ = service.UpdatePHPPath(currentPath, true)
	}
	if category == "nodejs" {
		_ = service.UpdateNodePath(currentPath, true)
	}
	if category == "python" {
		_ = service.UpdatePythonPath(currentPath, true)
	}
	if wasRunning {
		return a.StartService(serviceName)
	}
	a.orchestrator.RequestRefresh()
	return nil
}

// StartAllServices starts the default stack (MySQL, PHP, Apache)
func (a *App) StartAllServices() error {
	_ = a.StartService("MySQL")
	_ = a.StartService("PHP")
	return a.StartService("Apache")
}

func (a *App) updateMySQLConfig(mysqlPath string, port int) error {
	dataDir := filepath.Join(mysqlPath, "data"); tmpDir := filepath.Join(mysqlPath, "tmp"); binDir := filepath.Join(mysqlPath, "bin"); iniPath := filepath.Join(mysqlPath, "my.ini")
	err := service.UpdateMySQLConfig(mysqlPath, dataDir, tmpDir, port); if err != nil { return err }
	return service.InitializeMySQLDataDir(binDir, mysqlPath, dataDir, iniPath)
}
func (a *App) updateApacheConfig(apachePath string, port int) error {
	if port <= 0 {
		port = 80
	}
	wwwDir := a.cfg.WWWRoot
	_ = os.MkdirAll(wwwDir, 0755)
	phpInfo := a.orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" {
		phpPort = phpInfo.Port
	}

	vhostsContent := ""
	sslDir := filepath.Join(config.GetBaseDir(), "ssl")
	for name, targetPort := range a.cfg.Proxies {
		vhostsContent += service.GenerateProxyVHost(name, targetPort, port, a.cfg.ApacheHTTPS, sslDir)
		_ = service.AddHostWithElevation("127.0.0.1", name+".test")
		if a.cfg.ApacheHTTPS {
			_ = ssl.SignCertificate(sslDir, name+".test", sslDir)
		}
	}

	return service.UpdateApacheConfig(apachePath, "", "", vhostsContent, port, wwwDir, phpPort, a.cfg.ApacheHTTPS)
}
func (a *App) updateNginxConfig(nginxPath string, port int) error {
	if port <= 0 {
		port = 80
	}
	wwwDir := a.cfg.WWWRoot
	_ = os.MkdirAll(wwwDir, 0755)
	phpInfo := a.orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" {
		phpPort = phpInfo.Port
	}

	var proxies []service.ProxyConfig
	sslDir := filepath.Join(config.GetBaseDir(), "ssl")
	for name, targetPort := range a.cfg.Proxies {
		proxies = append(proxies, service.ProxyConfig{Name: name, TargetPort: targetPort})
		_ = service.AddHostWithElevation("127.0.0.1", name+".test")
		if a.cfg.NginxHTTPS {
			_ = ssl.SignCertificate(sslDir, name+".test", sslDir)
		}
	}

	return service.UpdateNginxConfig(nginxPath, wwwDir, phpPort, port, a.cfg.NginxHTTPS, proxies)
}

// StopAllServices stops all currently running background services
func (a *App) StopAllServices() { a.orchestrator.StopAll() }

// OpenTerminal opens a terminal at the current server root directory
func (a *App) OpenTerminal(terminalType string) {
	a.OpenTerminalAtPath(terminalType, a.cfg.WWWRoot)
}

// OpenTerminalAtPath opens a terminal at a specific local path with the Ostenia environment variables set
func (a *App) OpenTerminalAtPath(terminalType string, path string) {
	_, _, phpPath := a.getPluginPaths("PHP")
	_, mysqlBinDir, mysqlCurrentPath := a.getPluginPaths("MySQL")
	mysqlPath := filepath.Join(mysqlCurrentPath, "bin")
	// Fallback if current doesn't exist
	if _, err := os.Stat(mysqlPath); os.IsNotExist(err) {
		_ = filepath.Walk(mysqlBinDir, func(p string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == exeMySQL {
				mysqlPath = filepath.Dir(p)
				return filepath.SkipDir
			}
			return nil
		})
	}
	_, _, nodePath := a.getPluginPaths("Node.js")

	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = "PATH=" + phpPath + ";" + mysqlPath + ";" + nodePath + ";" + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+phpPath+";"+mysqlPath+";"+nodePath)
	}
	cmd := service.NewTerminal(path, env)
	cmd.Open(terminalType)
}

// DeleteVersion deletes a specific version folder of a plugin
func (a *App) DeleteVersion(serviceName string, version string) error {
	return a.downloader.DeleteVersion(serviceName, version)
}

// SetApacheHTTPS enables or disables HTTPS support for Apache
func (a *App) SetApacheHTTPS(enabled bool) error {
	a.cfg.ApacheHTTPS = enabled; err := config.SaveConfig(a.cfg); if err != nil { return err }
	if a.orchestrator.IsRunning("Apache") {
		_ = a.StopService("Apache")
		time.Sleep(600 * time.Millisecond)
		return a.StartService("Apache")
	}
	return nil
}

// SetNginxHTTPS enables or disables HTTPS support for Nginx
func (a *App) SetNginxHTTPS(enabled bool) error {
	a.cfg.NginxHTTPS = enabled; err := config.SaveConfig(a.cfg); if err != nil { return err }
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.StopService("Nginx")
		time.Sleep(600 * time.Millisecond)
		return a.StartService("Nginx")
	}
	return nil
}

// OpenServiceTerminal opens a terminal at the binary directory of a specific service
func (a *App) OpenServiceTerminal(serviceName string, terminalType string) error {
	category, binDir, _ := a.getPluginPaths(serviceName)
	targetDir := a.getServiceTargetDir(category, binDir)

	if targetDir == "" {
		targetDir = binDir
	}
	a.OpenTerminalAtPath(terminalType, targetDir)
	return nil
}

func (a *App) getServiceTargetDir(category string, binDir string) string {
	exeMap := map[string]string{
		"nginx":  exeNginx,
		"apache": exeApache,
		"mysql":  exeMySQL,
		"nodejs": exeNode,
		"python": exePython,
	}

	exeName, ok := exeMap[category]
	if !ok {
		return binDir
	}

	var targetDir string
	_ = filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == exeName {
			targetDir = filepath.Dir(path)
			return filepath.SkipDir
		}
		return nil
	})
	return targetDir
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
	if err != nil {
		return apps
	}

	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			port := 0
			if p, ok := a.cfg.Proxies[name]; ok {
				port = p
			}
			apps = append(apps, ProxyAppInfo{Name: name, Port: port})
		}
	}
	return apps
}

// SaveProxyPort saves the proxy port for a specific folder and reconfigures web servers
func (a *App) SaveProxyPort(name string, port int) error {
	if a.cfg.Proxies == nil {
		a.cfg.Proxies = make(map[string]int)
	}
	if port <= 0 {
		delete(a.cfg.Proxies, name)
	} else {
		a.cfg.Proxies[name] = port
	}
	err := config.SaveConfig(a.cfg)
	if err != nil {
		return err
	}

	// Trigger web server re-config
	if a.orchestrator.IsRunning("Apache") {
		_ = a.updateApacheConfig(filepath.Join(config.GetBaseDir(), "bin", "apache", "current"), a.orchestrator.GetDetailedInfo("Apache").Port)
		_ = a.StopService("Apache")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.updateNginxConfig(filepath.Join(config.GetBaseDir(), "bin", "nginx", "current"), a.orchestrator.GetDetailedInfo("Nginx").Port)
		_ = a.StopService("Nginx")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Nginx")
	}

	return nil
}
