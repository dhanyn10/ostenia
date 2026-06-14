package main

import (
	"context"
	"encoding/json"
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

// App struct
type App struct {
	ctx          context.Context
	downloader   *plugins.Manager
	orchestrator *service.Orchestrator
	symlinkMgr   *service.SymlinkManager
	sshManager   *service.SSHManager
	cfg          *config.Config
}

// NewApp creates a new App struct
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

	cfg, _ := config.LoadConfig()
	a.cfg = cfg

	// Initial setup of directories in current base dir
	a.ensureEnvironmentStructure()

	// Start the periodic watcher for services
	a.orchestrator.StartWatcher()

	// Start proxy port watcher
	go a.startProxyWatcher()
}

func (a *App) SelectDefaultEditor() (string, error) {
	selected, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Select Default Text Editor",
		Filters: []wruntime.FileFilter{
			{DisplayName: "Executables (*.exe;*.app)", Pattern: "*.exe;*.app"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err == nil && selected != "" {
		a.cfg.DefaultEditor = selected
		config.SaveConfig(a.cfg)
	}
	return selected, err
}

func (a *App) SetDefaultEditor(editor string) error {
	a.cfg.DefaultEditor = editor
	return config.SaveConfig(a.cfg)
}

func (a *App) ToggleDevTools() {
	wruntime.WindowExecJS(a.ctx, "window.runtime.WindowToggleDevTools()")
}

func (a *App) Minimize() { wruntime.WindowMinimise(a.ctx) }
func (a *App) Maximize() { wruntime.WindowMaximise(a.ctx) }
func (a *App) Unmaximize() { wruntime.WindowUnmaximise(a.ctx) }
func (a *App) Close() { wruntime.Quit(a.ctx) }

type ProfileData struct {
	Config      *config.Config      `json:"config,omitempty"`
	SSHSessions []config.SSHSession `json:"sshSessions,omitempty"`
}

func (a *App) ExportProfile(includeConfig bool, includeSSH bool) error {
	profile := ProfileData{}
	if includeConfig {
		profile.Config = a.cfg
	}
	if includeSSH {
		sessions, err := config.LoadSSHSessions()
		if err == nil {
			profile.SSHSessions = sessions
		}
	}

	filePath, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Export Ostenia Profile",
		DefaultFilename: "ostenia_profile.json",
		Filters: []wruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || filePath == "" {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (a *App) ImportProfile() error {
	filePath, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Import Ostenia Profile",
		Filters: []wruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || filePath == "" {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var profile ProfileData
	err = json.Unmarshal(data, &profile)
	if err != nil {
		return err
	}

	if profile.Config != nil {
		// We preserve BaseDir and WWWRoot to avoid breaking the current installation
		profile.Config.BaseDir = a.cfg.BaseDir
		profile.Config.WWWRoot = a.cfg.WWWRoot
		a.cfg = profile.Config
		config.SaveConfig(a.cfg)
	}

	if profile.SSHSessions != nil {
		config.SaveSSHSessions(profile.SSHSessions)
	}

	wruntime.EventsEmit(a.ctx, "environment_changed", a.cfg)
	return nil
}

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
	for _, dir := range dirs { if _, err := os.Stat(dir); os.IsNotExist(err) { os.MkdirAll(dir, 0755) } }
	if a.cfg != nil && a.cfg.WWWRoot != "" { os.MkdirAll(a.cfg.WWWRoot, 0755) }
}

func (a *App) UpdateActiveTab(tab string) { a.orchestrator.SetActiveTab(tab) }
func (a *App) GetConfig() *config.Config { return a.cfg }
func (a *App) IsAdmin() bool { return service.IsAdmin() }
func (a *App) GetPrerequisites() []plugins.DownloadTask { return plugins.GetLatestKnownVersions() }
func (a *App) GetServiceStatus(serviceName string) service.ServiceDetailedInfo { return a.orchestrator.GetDetailedInfo(serviceName) }

func (a *App) OpenHeidiSQL() error {
	exePath, _ := plugins.DetectHeidiSQLInstallation()
	if exePath == "" {
		return fmt.Errorf("HeidiSQL is not installed")
	}
	cmd := exec.Command("cmd", "/c", "start", "", exePath)
	plugins_utils.SetHideWindow(cmd)
	return cmd.Run()
}

func (a *App) SetWWWRoot(path string) error {
	fmt.Printf("[App] Setting Server Root (www) to: %s\n", path)
	a.cfg.WWWRoot = path
	err := config.SaveConfig(a.cfg)
	if err != nil { return err }
	os.MkdirAll(path, 0755)
	if a.orchestrator.IsRunning("Apache") { a.StopService("Apache"); time.Sleep(500 * time.Millisecond); a.StartService("Apache") }
	if a.orchestrator.IsRunning("Nginx") { a.StopService("Nginx"); time.Sleep(500 * time.Millisecond); a.StartService("Nginx") }
	return nil
}

type ProxyStatusInfo struct {
	Name   string `json:"name"`
	IsUp   bool   `json:"isUp"`
	Port   int    `json:"port"`
}

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

func (a *App) OpenProxyTerminal(name string, terminalType string) error {
	path := filepath.Join(a.cfg.WWWRoot, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("folder %s not found", name)
	}
	a.OpenTerminalAtPath(terminalType, path)
	return nil
}

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

func (a *App) SelectServerRoot() (string, error) {
	selectedDir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Ostenia Apps Location"})
	if err != nil { return "", err }
	if selectedDir != "" { err = a.SetServerRoot(selectedDir); if err != nil { return "", err } }
	return selectedDir, nil
}

func (a *App) SelectWWWRoot() (string, error) {
	selectedDir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Server Root (www)"})
	if err != nil { return "", err }
	if selectedDir != "" { err = a.SetWWWRoot(selectedDir); if err != nil { return "", err } }
	return selectedDir, nil
}

func (a *App) OpenServerRootFolder() error { return service.OpenExplorer(a.cfg.WWWRoot) }
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

func (a *App) OpenPluginFolder(serviceName string) error {
	_, binDir, _ := a.getPluginPaths(serviceName)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		os.MkdirAll(binDir, 0755)
	}
	return service.OpenExplorer(binDir)
}

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

func (a *App) CancelDownload(taskName string) { a.downloader.CancelDownload(taskName) }

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
	var binPath, basePath string
	filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == exeName {
			binPath = path
			basePath = filepath.Dir(filepath.Dir(path))
			if exeName == "nginx.exe" {
				basePath = filepath.Dir(path)
			}
			return filepath.SkipDir
		}
		return nil
	})
	return binPath, basePath
}

func (a *App) startMySQLService(binDir string) error {
	mysqlBin, mysqlBase := a.findExecutable(binDir, "mysqld.exe")
	if mysqlBin == "" {
		return fmt.Errorf("mysqld.exe not found")
	}
	port := a.orchestrator.GetDetailedInfo("MySQL").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{3306, 3307, 3308})
		if err != nil {
			return err
		}
		port = p
	}
	os.Setenv("PATH", filepath.Dir(mysqlBin)+";"+os.Getenv("PATH"))
	if err := a.updateMySQLConfig(mysqlBase, port); err != nil {
		return err
	}
	iniPath := filepath.Join(mysqlBase, "my.ini")
	return a.orchestrator.StartServiceWithPort("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath, "--console"}, filepath.Dir(mysqlBin), port)
}

func (a *App) startApacheService(binDir string) error {
	if a.orchestrator.IsRunning("Nginx") {
		a.StopService("Nginx")
		time.Sleep(600 * time.Millisecond)
	}
	apacheBin, apacheBase := a.findExecutable(binDir, "httpd.exe")
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
	os.Setenv("PATH", phpPath+";"+os.Getenv("PATH")+";"+nodePath)
	if err := a.updateApacheConfig(apacheBase, port); err != nil {
		return err
	}
	return a.orchestrator.StartServiceWithPort("Apache", apacheBin, []string{}, apacheBase, port)
}

func (a *App) startNginxService(binDir string) error {
	if a.orchestrator.IsRunning("Apache") {
		a.StopService("Apache")
		time.Sleep(600 * time.Millisecond)
	}
	nginxBin, nginxBase := a.findExecutable(binDir, "nginx.exe")
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
	phpCgi := filepath.Join(currentPath, "php-cgi.exe")
	if _, err := os.Stat(phpCgi); os.IsNotExist(err) {
		return fmt.Errorf("php-cgi.exe not found")
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
	os.Setenv("PATH", currentPath+";"+os.Getenv("PATH"))
	_ = service.UpdatePHPPath(currentPath, true)
	os.Setenv("PHP_FCGI_MAX_REQUESTS", "1000")
	err := a.orchestrator.StartServiceWithPort("PHP", phpCgi, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, currentPath, port)
	if err == nil {
		a.restartDependentWebServers()
	}
	return err
}

func (a *App) restartDependentWebServers() {
	for _, srv := range []string{"Apache", "Nginx"} {
		if a.orchestrator.IsRunning(srv) {
			a.StopService(srv)
			time.Sleep(600 * time.Millisecond)
			a.StartService(srv)
		}
	}
}

func (a *App) StopService(serviceName string) {
	_, _, currentPath := a.getPluginPaths(serviceName)

	if serviceName == "OpenSSL" {
		baseDir := config.GetBaseDir()
		caDir := filepath.Join(baseDir, "ssl")
		os.RemoveAll(caDir)
		os.MkdirAll(caDir, 0755)
		a.SetApacheHTTPS(false)
		a.SetNginxHTTPS(false)
		a.orchestrator.RequestRefresh()
		return
	}
	if serviceName == "PHP" {
		_ = service.UpdatePHPPath(currentPath, false)
	}
	if serviceName == "Node.js" {
		_ = service.UpdateNodePath(currentPath, false)
		a.orchestrator.RequestRefresh()
		return
	}
	if serviceName == "Python" {
		_ = service.UpdatePythonPath(currentPath, false)
		a.orchestrator.RequestRefresh()
		return
	}
	a.orchestrator.StopService(serviceName)
}

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
		a.StopService(serviceName)
		time.Sleep(600 * time.Millisecond)
	}
	os.Remove(currentPath)
	if _, err := os.Stat(targetDir); err == nil {
		cmd := exec.Command("cmd", "/c", "mklink", "/J", currentPath, targetDir)
		plugins_utils.SetHideWindow(cmd)
		_ = cmd.Run()
	}
	if category == "php" {
		os.Setenv("PATH", currentPath+";"+os.Getenv("PATH"))
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

func (a *App) StartAllServices() error { a.StartService("MySQL"); a.StartService("PHP"); return a.StartService("Apache") }
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
	os.MkdirAll(wwwDir, 0755)
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
	os.MkdirAll(wwwDir, 0755)
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
func (a *App) StopAllServices() { a.orchestrator.StopAll() }
func (a *App) OpenTerminal(terminalType string) {
	a.OpenTerminalAtPath(terminalType, a.cfg.WWWRoot)
}

func (a *App) OpenTerminalAtPath(terminalType string, path string) {
	_, _, phpPath := a.getPluginPaths("PHP")
	_, mysqlBinDir, mysqlCurrentPath := a.getPluginPaths("MySQL")
	mysqlPath := filepath.Join(mysqlCurrentPath, "bin")
	// Fallback if current doesn't exist
	if _, err := os.Stat(mysqlPath); os.IsNotExist(err) {
		filepath.Walk(mysqlBinDir, func(p string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "mysqld.exe" {
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
func (a *App) DeleteVersion(serviceName string, version string) error { return a.downloader.DeleteVersion(serviceName, version) }
func (a *App) SetApacheHTTPS(enabled bool) error {
	a.cfg.ApacheHTTPS = enabled; err := config.SaveConfig(a.cfg); if err != nil { return err }
	if a.orchestrator.IsRunning("Apache") { a.StopService("Apache"); time.Sleep(600 * time.Millisecond); return a.StartService("Apache") }
	return nil
}
func (a *App) SetNginxHTTPS(enabled bool) error {
	a.cfg.NginxHTTPS = enabled; err := config.SaveConfig(a.cfg); if err != nil { return err }
	if a.orchestrator.IsRunning("Nginx") { a.StopService("Nginx"); time.Sleep(600 * time.Millisecond); return a.StartService("Nginx") }
	return nil
}
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
		"nginx":  "nginx.exe",
		"apache": "httpd.exe",
		"mysql":  "mysqld.exe",
		"nodejs": "node.exe",
		"python": "python.exe",
	}

	exeName, ok := exeMap[category]
	if !ok {
		return binDir
	}

	var targetDir string
	filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == exeName {
			targetDir = filepath.Dir(path)
			return filepath.SkipDir
		}
		return nil
	})
	return targetDir
}
func (a *App) GetPHPExtensions() ([]service.PHPExtensionInfo, error) { baseDir := config.GetBaseDir(); phpPath := filepath.Join(baseDir, "bin", "php", "current"); return service.GetPHPExtensions(phpPath) }
func (a *App) TogglePHPExtension(extName string, enable bool) error {
	baseDir := config.GetBaseDir(); phpPath := filepath.Join(baseDir, "bin", "php", "current"); err := service.TogglePHPExtension(phpPath, extName, enable); if err != nil { return err }
	if a.orchestrator.IsRunning("PHP") { a.StopService("PHP"); time.Sleep(600 * time.Millisecond); return a.StartService("PHP") }
	return nil
}

func (a *App) GetSSHSessions() ([]config.SSHSession, error) {
	return config.LoadSSHSessions()
}

func (a *App) SaveSSHSessions(sessions []config.SSHSession) error {
	return config.SaveSSHSessions(sessions)
}

func (a *App) AddSSHSession(session config.SSHSession) error {
	return config.AddSSHSession(session)
}

func (a *App) UpdateSSHSession(session config.SSHSession) error {
	return config.UpdateSSHSession(session)
}

func (a *App) DeleteSSHSession(id string) error {
	return config.DeleteSSHSession(id)
}

func (a *App) ConnectSSH(session config.SSHSession) error {
	return a.sshManager.Connect(session)
}

func (a *App) DisconnectSSH(sessionID string) {
	a.sshManager.Disconnect(sessionID)
}

func (a *App) SendSSHInput(sessionID string, data string) error {
	return a.sshManager.SendInput(sessionID, data)
}

func (a *App) ResizeSSHTerminal(sessionID string, cols int, rows int) error {
	return a.sshManager.ResizeTerminal(sessionID, cols, rows)
}

func (a *App) GetRemoteFiles(sessionID string, path string) ([]service.RemoteFile, error) {
	return a.sshManager.ListFiles(sessionID, path)
}

func (a *App) ExecuteSFTPAction(sessionID string, action string, path string, target string) error {
	return a.sshManager.ExecuteSFTPAction(sessionID, action, path, target)
}

func (a *App) EditRemoteFile(sessionID string, remotePath string) error {
	return a.sshManager.EditFile(sessionID, remotePath, a.cfg.DefaultEditor)
}

func (a *App) GetRemoteCurrentPath(sessionID string) (string, error) {
	return a.sshManager.GetCurrentPath(sessionID)
}

func (a *App) DownloadRemoteFile(sessionID string, remotePath string) error {
	fileName := filepath.Base(remotePath)
	localPath, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title: "Download File",
		DefaultFilename: fileName,
	})
	if err != nil || localPath == "" {
		return err
	}
	return a.sshManager.DownloadFile(sessionID, remotePath, localPath)
}

func (a *App) UploadRemoteFile(sessionID string, remoteDir string) error {
	localPath, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Upload File",
	})
	if err != nil || localPath == "" {
		return err
	}
	remotePath := filepath.ToSlash(filepath.Join(remoteDir, filepath.Base(localPath)))
	return a.sshManager.UploadFile(sessionID, localPath, remotePath)
}

type ProxyAppInfo struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

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
		a.updateApacheConfig(filepath.Join(config.GetBaseDir(), "bin", "apache", "current"), a.orchestrator.GetDetailedInfo("Apache").Port)
		// Apache might need full restart for some changes, but usually reload is enough if handled in updateApacheConfig
		// For now we follow the existing pattern in SetWWWRoot
		a.StopService("Apache")
		time.Sleep(500 * time.Millisecond)
		a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		a.updateNginxConfig(filepath.Join(config.GetBaseDir(), "bin", "nginx", "current"), a.orchestrator.GetDetailedInfo("Nginx").Port)
		a.StopService("Nginx")
		time.Sleep(500 * time.Millisecond)
		a.StartService("Nginx")
	}

	return nil
}
