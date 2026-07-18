package backend

import (
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/network"
	"ostenia/internal/plugins"
	plugins_utils "ostenia/internal/plugins/utils"
	"ostenia/internal/service"
	"ostenia/internal/backend/interfaces"
	"path/filepath"
	"time"
)

const serviceNodeJS = "Node.js"

// GetServiceStatus returns detailed information about a specific service
func (a *App) GetServiceStatus(serviceName string) interfaces.ServiceDetailedInfo {
	return a.orchestrator.GetDetailedInfo(serviceName)
}

// StartService starts a background service by name
func (a *App) StartService(serviceName string) error {
	_, binDir, currentPath := a.getPluginPaths(serviceName)
	fmt.Printf("[App] Starting service: %s\n", serviceName)

	switch serviceName {
	case serviceNodeJS:
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
	if serviceName == serviceNodeJS {
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
	if err := a.sslManager.GenerateRootCA(caDir); err != nil {
		return err
	}
	a.orchestrator.RequestRefresh()
	return nil
}

func (a *App) findExecutable(binDir string, exeName string) (string, string) {
	currentPath := filepath.Join(binDir, "current")

	// 1. Try "current" link first for efficiency
	if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		if bin, base := a.checkStandardExePath(resolved, exeName); bin != "" {
			return bin, base
		}
	}

	// 2. Fallback to Walk if "current" is not valid or doesn't match
	return a.walkForExecutable(binDir, exeName)
}

func (a *App) checkStandardExePath(resolved string, exeName string) (string, string) {
	// Standard bin path
	path := filepath.Join(resolved, "bin", exeName)
	if exeName == exeNginx {
		path = filepath.Join(resolved, exeName)
	}

	if _, err := os.Stat(path); err == nil {
		return path, resolved
	}

	// Apache specific fallback
	if exeName == exeApache {
		apachePath := filepath.Join(resolved, "Apache24", "bin", exeName)
		if _, err := os.Stat(apachePath); err == nil {
			return apachePath, filepath.Join(resolved, "Apache24")
		}
	}

	return "", ""
}

func (a *App) walkForExecutable(binDir string, exeName string) (string, string) {
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
	_ = os.Setenv("PATH", filepath.Dir(mysqlBin)+";"+os.Getenv("PATH")) // NOSONAR
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
	_, _, nodePath := a.getPluginPaths(serviceNodeJS)
	_ = os.Setenv("PATH", phpPath+";"+os.Getenv("PATH")+";"+nodePath) // NOSONAR
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
	_ = os.Setenv("PATH", currentPath+";"+os.Getenv("PATH")) // NOSONAR
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

// StartAllServices starts the default stack (MySQL, PHP, Apache)
func (a *App) StartAllServices() error {
	_ = a.StartService("MySQL")
	_ = a.StartService("PHP")
	return a.StartService("Apache")
}

// StopAllServices stops all currently running background services
func (a *App) StopAllServices() { a.orchestrator.StopAll() }

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
			_ = a.sslManager.SignCertificate(sslDir, name+".test", sslDir)
		}
	}

	return service.UpdateApacheConfig(apachePath, vhostsContent, port, wwwDir, phpPort, a.cfg.ApacheHTTPS)
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
			_ = a.sslManager.SignCertificate(sslDir, name+".test", sslDir)
		}
	}

	return service.UpdateNginxConfig(nginxPath, wwwDir, phpPort, port, a.cfg.NginxHTTPS, proxies)
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

// OpenHeidiSQL launches the HeidiSQL application if installed
func (a *App) OpenHeidiSQL() error {
	exePath, _ := plugins.DetectHeidiSQLInstallation()
	if exePath == "" {
		return fmt.Errorf("HeidiSQL is not installed")
	}
	cmdPath := filepath.Join(plugins_utils.GetSystemDirectory(), "cmd.exe")
	cmd := plugins_utils.Executor.Command(cmdPath, "/c", "start", "", exePath)
	cmd.Env = plugins_utils.SafeEnv()
	plugins_utils.SetHideWindow(cmd)
	return cmd.Run()
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
