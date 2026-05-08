package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/plugins/php"
	"ostenia/internal/plugins/python"
	"ostenia/internal/network"
	"ostenia/internal/service"
	"ostenia/internal/ssl"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	downloader   *plugins.Manager
	orchestrator *service.Orchestrator
	symlinkMgr   *service.SymlinkManager
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

	cfg, _ := config.LoadConfig()
	a.cfg = cfg

	// Initial setup of directories in current base dir
	a.ensureEnvironmentStructure()

	// Start the periodic watcher for services
	a.orchestrator.StartWatcher()

	// Start proxy port watcher
	go a.startProxyWatcher()
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
	baseDir := config.GetBaseDir()
	_, binDir, currentPath := a.getPluginPaths(serviceName)
	fmt.Printf("[App] Starting service: %s\n", serviceName)

	switch serviceName {
	case "Node.js":
		if _, err := os.Stat(currentPath); os.IsNotExist(err) {
			return fmt.Errorf("node.js not installed")
		}
		err := service.UpdateNodePath(currentPath, true)
		if err != nil {
			return err
		}
		a.orchestrator.RequestRefresh()
		return nil

	case "Python":
		if _, err := os.Stat(currentPath); os.IsNotExist(err) {
			return fmt.Errorf("python not installed")
		}
		err := service.UpdatePythonPath(currentPath, true)
		if err != nil {
			return err
		}
		a.orchestrator.RequestRefresh()
		return nil

	case "OpenSSL":
		caDir := filepath.Join(baseDir, "ssl")
		err := ssl.GenerateRootCA(caDir)
		if err != nil {
			return err
		}
		a.orchestrator.RequestRefresh()
		return nil

	case "MySQL":
		var mysqlBin string
		var mysqlBase string
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "mysqld.exe" {
				mysqlBin = path
				mysqlBase = filepath.Dir(filepath.Dir(path))
				return filepath.SkipDir
			}
			return nil
		})
		if mysqlBin == "" {
			return fmt.Errorf("mysqld.exe not found")
		}
		currentInfo := a.orchestrator.GetDetailedInfo("MySQL")
		port := currentInfo.Port
		if port <= 0 {
			p, err := network.GetAvailablePort([]int{3306, 3307, 3308})
			if err != nil {
				return err
			}
			port = p
		}
		os.Setenv("PATH", filepath.Dir(mysqlBin)+";"+os.Getenv("PATH"))
		err := a.updateMySQLConfig(mysqlBase, port)
		if err != nil {
			return err
		}
		iniPath := filepath.Join(mysqlBase, "my.ini")
		return a.orchestrator.StartServiceWithPort("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath, "--console"}, filepath.Dir(mysqlBin), port)

	case "Apache":
		if a.orchestrator.IsRunning("Nginx") {
			a.StopService("Nginx")
			time.Sleep(600 * time.Millisecond)
		}
		var apacheBase string
		var apacheBin string
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "httpd.exe" {
				apacheBin = path
				apacheBase = filepath.Dir(filepath.Dir(path))
				return filepath.SkipDir
			}
			return nil
		})
		if apacheBase == "" {
			return fmt.Errorf("apache installation not found")
		}
		currentInfo := a.orchestrator.GetDetailedInfo("Apache")
		port := currentInfo.Port
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
		err := a.updateApacheConfig(apacheBase, port)
		if err != nil {
			return err
		}
		return a.orchestrator.StartServiceWithPort("Apache", apacheBin, []string{}, apacheBase, port)

	case "Nginx":
		if a.orchestrator.IsRunning("Apache") {
			a.StopService("Apache")
			time.Sleep(600 * time.Millisecond)
		}
		var nginxBase string
		var nginxBin string
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			if !info.IsDir() && info.Name() == "nginx.exe" {
				nginxBin = path
				nginxBase = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
		if nginxBase == "" {
			return fmt.Errorf("nginx installation not found")
		}
		currentInfo := a.orchestrator.GetDetailedInfo("Nginx")
		port := currentInfo.Port
		if port <= 0 {
			p, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
			if err != nil {
				return err
			}
			port = p
		}
		err := a.updateNginxConfig(nginxBase, port)
		if err != nil {
			return err
		}
		return a.orchestrator.StartServiceWithPort("Nginx", nginxBin, []string{"-p", nginxBase}, nginxBase, port)

	case "HeidiSQL":
		hBin := filepath.Join(currentPath, "heidisql.exe")
		if _, err := os.Stat(hBin); os.IsNotExist(err) {
			hBin = filepath.Join(binDir, "heidisql.exe")
			if _, err := os.Stat(hBin); os.IsNotExist(err) {
				return fmt.Errorf("heidisql.exe not found")
			}
		}
		return a.orchestrator.StartService("HeidiSQL", hBin, []string{}, filepath.Dir(hBin))

	case "PHP":
		phpCgi := filepath.Join(currentPath, "php-cgi.exe")
		if _, err := os.Stat(phpCgi); os.IsNotExist(err) {
			return fmt.Errorf("php-cgi.exe not found")
		}
		err := service.UpdatePHPConfig(currentPath)
		if err != nil {
			fmt.Printf("[PHP] Warning: %v\n", err)
		}
		currentInfo := a.orchestrator.GetDetailedInfo("PHP")
		port := currentInfo.Port
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
		err = a.orchestrator.StartServiceWithPort("PHP", phpCgi, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, currentPath, port)
		if err == nil {
			if a.orchestrator.IsRunning("Apache") {
				a.StopService("Apache")
				time.Sleep(600 * time.Millisecond)
				a.StartService("Apache")
			}
			if a.orchestrator.IsRunning("Nginx") {
				a.StopService("Nginx")
				time.Sleep(600 * time.Millisecond)
				a.StartService("Nginx")
			}
		}
		return err
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
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
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		}
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
	var targetDir string

	switch category {
	case "nginx":
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "nginx.exe" {
				targetDir = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
	case "apache":
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "httpd.exe" {
				targetDir = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
	case "mysql":
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "mysqld.exe" {
				targetDir = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
	case "nodejs":
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "node.exe" {
				targetDir = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
	case "python":
		filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "python.exe" {
				targetDir = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})
	default:
		targetDir = binDir
	}

	if targetDir == "" {
		targetDir = binDir
	}
	a.OpenTerminalAtPath(terminalType, targetDir)
	return nil
}
func (a *App) GetPHPExtensions() ([]service.PHPExtensionInfo, error) { baseDir := config.GetBaseDir(); phpPath := filepath.Join(baseDir, "bin", "php", "current"); return service.GetPHPExtensions(phpPath) }
func (a *App) TogglePHPExtension(extName string, enable bool) error {
	baseDir := config.GetBaseDir(); phpPath := filepath.Join(baseDir, "bin", "php", "current"); err := service.TogglePHPExtension(phpPath, extName, enable); if err != nil { return err }
	if a.orchestrator.IsRunning("PHP") { a.StopService("PHP"); time.Sleep(600 * time.Millisecond); return a.StartService("PHP") }
	return nil
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
