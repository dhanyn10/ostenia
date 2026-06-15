package backend

import (
	"context"
	"fmt"
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

const (
	exeNginx  = "nginx.exe"
	exeApache = "httpd.exe"
	exeMySQL  = "mysqld.exe"
	exePHP    = "php-cgi.exe"
	exeNode   = "node.exe"
	exePython = "python.exe"
)

type ServiceManager struct {
	Ctx          context.Context
	Downloader   *plugins.Manager
	Orchestrator *service.Orchestrator
	SymlinkMgr   *service.SymlinkManager
	Cfg          *config.Config
}

func (s *ServiceManager) GetPluginPaths(serviceName string) (category string, binDir string, currentPath string) {
	category = strings.ToLower(serviceName)
	if category == "node.js" {
		category = "nodejs"
	}
	binDir = filepath.Join(config.GetBaseDir(), "bin", category)
	currentPath = filepath.Join(binDir, "current")
	return
}

func (s *ServiceManager) StartService(serviceName string) error {
	_, binDir, currentPath := s.GetPluginPaths(serviceName)
	fmt.Printf("[ServiceManager] Starting service: %s\n", serviceName)

	switch serviceName {
	case "Node.js":
		return s.startNodeService(currentPath)
	case "Python":
		return s.startPythonService(currentPath)
	case "OpenSSL":
		return s.startOpenSSLService()
	case "MySQL":
		return s.startMySQLService(binDir)
	case "Apache":
		return s.startApacheService(binDir)
	case "Nginx":
		return s.startNginxService(binDir)
	case "HeidiSQL":
		return s.startHeidiSQLService()
	case "PHP":
		return s.startPHPService(currentPath)
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

func (s *ServiceManager) startNodeService(currentPath string) error {
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("node.js not installed")
	}
	if err := service.UpdateNodePath(currentPath, true); err != nil {
		return err
	}
	s.Orchestrator.RequestRefresh()
	return nil
}

func (s *ServiceManager) startPythonService(currentPath string) error {
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("python not installed")
	}
	if err := service.UpdatePythonPath(currentPath, true); err != nil {
		return err
	}
	s.Orchestrator.RequestRefresh()
	return nil
}

func (s *ServiceManager) startOpenSSLService() error {
	caDir := filepath.Join(config.GetBaseDir(), "ssl")
	if err := ssl.GenerateRootCA(caDir); err != nil {
		return err
	}
	s.Orchestrator.RequestRefresh()
	return nil
}

func (s *ServiceManager) findExecutable(binDir string, exeName string) (string, string) {
	currentPath := filepath.Join(binDir, "current")
	if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		path := filepath.Join(resolved, "bin", exeName)
		if exeName == exeNginx {
			path = filepath.Join(resolved, exeName)
		}
		if _, err := os.Stat(path); err == nil {
			return path, resolved
		}
		if exeName == exeApache {
			path = filepath.Join(resolved, "Apache24", "bin", exeName)
			if _, err := os.Stat(path); err == nil {
				return path, filepath.Join(resolved, "Apache24")
			}
		}
	}
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

func (s *ServiceManager) startMySQLService(binDir string) error {
	mysqlBin, mysqlBase := s.findExecutable(binDir, exeMySQL)
	if mysqlBin == "" {
		return fmt.Errorf("%s not found", exeMySQL)
	}
	port := s.Orchestrator.GetDetailedInfo("MySQL").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{3306, 3307, 3308})
		if err != nil {
			return err
		}
		port = p
	}
	_ = os.Setenv("PATH", filepath.Dir(mysqlBin)+";"+os.Getenv("PATH"))
	if err := s.UpdateMySQLConfig(mysqlBase, port); err != nil {
		return err
	}
	iniPath := filepath.Join(mysqlBase, "my.ini")
	return s.Orchestrator.StartServiceWithPort("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath, "--console"}, filepath.Dir(mysqlBin), port)
}

func (s *ServiceManager) startApacheService(binDir string) error {
	if s.Orchestrator.IsRunning("Nginx") {
		_ = s.StopService("Nginx")
		time.Sleep(600 * time.Millisecond)
	}
	apacheBin, apacheBase := s.findExecutable(binDir, exeApache)
	if apacheBase == "" {
		return fmt.Errorf("apache installation not found")
	}
	port := s.Orchestrator.GetDetailedInfo("Apache").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil {
			return err
		}
		port = p
	}
	_, _, phpPath := s.GetPluginPaths("PHP")
	_, _, nodePath := s.GetPluginPaths("Node.js")
	_ = os.Setenv("PATH", phpPath+";"+os.Getenv("PATH")+";"+nodePath)
	if err := s.UpdateApacheConfig(apacheBase, port); err != nil {
		return err
	}
	return s.Orchestrator.StartServiceWithPort("Apache", apacheBin, []string{}, apacheBase, port)
}

func (s *ServiceManager) startNginxService(binDir string) error {
	if s.Orchestrator.IsRunning("Apache") {
		_ = s.StopService("Apache")
		time.Sleep(600 * time.Millisecond)
	}
	nginxBin, nginxBase := s.findExecutable(binDir, exeNginx)
	if nginxBase == "" {
		return fmt.Errorf("nginx installation not found")
	}
	port := s.Orchestrator.GetDetailedInfo("Nginx").Port
	if port <= 0 {
		p, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil {
			return err
		}
		port = p
	}
	if err := s.UpdateNginxConfig(nginxBase, port); err != nil {
		return err
	}
	return s.Orchestrator.StartServiceWithPort("Nginx", nginxBin, []string{"-p", nginxBase}, nginxBase, port)
}

func (s *ServiceManager) startHeidiSQLService() error {
	exePath, _ := plugins.DetectHeidiSQLInstallation()
	if exePath != "" {
		s.Orchestrator.RequestRefresh()
		return nil
	}
	tasks := plugins.GetLatestKnownVersions()
	for _, t := range tasks {
		if t.Name == "HeidiSQL" {
			return s.Downloader.DownloadAndExtract(t)
		}
	}
	return fmt.Errorf("HeidiSQL task not found")
}

func (s *ServiceManager) startPHPService(currentPath string) error {
	phpCgi := filepath.Join(currentPath, exePHP)
	if _, err := os.Stat(phpCgi); os.IsNotExist(err) {
		return fmt.Errorf("%s not found", exePHP)
	}
	_ = service.UpdatePHPConfig(currentPath)
	port := s.Orchestrator.GetDetailedInfo("PHP").Port
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
	err := s.Orchestrator.StartServiceWithPort("PHP", phpCgi, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, currentPath, port)
	if err == nil {
		s.RestartDependentWebServers()
	}
	return err
}

func (s *ServiceManager) RestartDependentWebServers() {
	for _, srv := range []string{"Apache", "Nginx"} {
		if s.Orchestrator.IsRunning(srv) {
			_ = s.StopService(srv)
			time.Sleep(600 * time.Millisecond)
			_ = s.StartService(srv)
		}
	}
}

func (s *ServiceManager) StopService(serviceName string) error {
	_, _, currentPath := s.GetPluginPaths(serviceName)

	if serviceName == "OpenSSL" {
		baseDir := config.GetBaseDir()
		caDir := filepath.Join(baseDir, "ssl")
		_ = os.RemoveAll(caDir)
		_ = os.MkdirAll(caDir, 0755)
		_ = s.SetApacheHTTPS(false)
		_ = s.SetNginxHTTPS(false)
		s.Orchestrator.RequestRefresh()
		return nil
	}
	if serviceName == "PHP" {
		_ = service.UpdatePHPPath(currentPath, false)
	}
	if serviceName == "Node.js" {
		_ = service.UpdateNodePath(currentPath, false)
		s.Orchestrator.RequestRefresh()
		return nil
	}
	if serviceName == "Python" {
		_ = service.UpdatePythonPath(currentPath, false)
		s.Orchestrator.RequestRefresh()
		return nil
	}
	return s.Orchestrator.StopService(serviceName)
}

func (s *ServiceManager) UpdateMySQLConfig(mysqlPath string, port int) error {
	dataDir := filepath.Join(mysqlPath, "data"); tmpDir := filepath.Join(mysqlPath, "tmp"); binDir := filepath.Join(mysqlPath, "bin"); iniPath := filepath.Join(mysqlPath, "my.ini")
	err := service.UpdateMySQLConfig(mysqlPath, dataDir, tmpDir, port); if err != nil { return err }
	return service.InitializeMySQLDataDir(binDir, mysqlPath, dataDir, iniPath)
}

func (s *ServiceManager) UpdateApacheConfig(apachePath string, port int) error {
	if port <= 0 { port = 80 }
	wwwDir := s.Cfg.WWWRoot
	_ = os.MkdirAll(wwwDir, 0755)
	phpInfo := s.Orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" { phpPort = phpInfo.Port }

	vhostsContent := ""
	sslDir := filepath.Join(config.GetBaseDir(), "ssl")
	for name, targetPort := range s.Cfg.Proxies {
		vhostsContent += service.GenerateProxyVHost(name, targetPort, port, s.Cfg.ApacheHTTPS, sslDir)
		_ = service.AddHostWithElevation("127.0.0.1", name+".test")
		if s.Cfg.ApacheHTTPS { _ = ssl.SignCertificate(sslDir, name+".test", sslDir) }
	}
	return service.UpdateApacheConfig(apachePath, "", "", vhostsContent, port, wwwDir, phpPort, s.Cfg.ApacheHTTPS)
}

func (s *ServiceManager) UpdateNginxConfig(nginxPath string, port int) error {
	if port <= 0 { port = 80 }
	wwwDir := s.Cfg.WWWRoot
	_ = os.MkdirAll(wwwDir, 0755)
	phpInfo := s.Orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" { phpPort = phpInfo.Port }

	var proxies []service.ProxyConfig
	sslDir := filepath.Join(config.GetBaseDir(), "ssl")
	for name, targetPort := range s.Cfg.Proxies {
		proxies = append(proxies, service.ProxyConfig{Name: name, TargetPort: targetPort})
		_ = service.AddHostWithElevation("127.0.0.1", name+".test")
		if s.Cfg.NginxHTTPS { _ = ssl.SignCertificate(sslDir, name+".test", sslDir) }
	}
	return service.UpdateNginxConfig(nginxPath, wwwDir, phpPort, port, s.Cfg.NginxHTTPS, proxies)
}

func (s *ServiceManager) SetApacheHTTPS(enabled bool) error {
	s.Cfg.ApacheHTTPS = enabled; err := config.SaveConfig(s.Cfg); if err != nil { return err }
	if s.Orchestrator.IsRunning("Apache") {
		_ = s.StopService("Apache")
		time.Sleep(600 * time.Millisecond)
		return s.StartService("Apache")
	}
	return nil
}

func (s *ServiceManager) SetNginxHTTPS(enabled bool) error {
	s.Cfg.NginxHTTPS = enabled; err := config.SaveConfig(s.Cfg); if err != nil { return err }
	if s.Orchestrator.IsRunning("Nginx") {
		_ = s.StopService("Nginx")
		time.Sleep(600 * time.Millisecond)
		return s.StartService("Nginx")
	}
	return nil
}

func (s *ServiceManager) InstallPrerequisite(task plugins.DownloadTask) error {
	err := s.Downloader.DownloadAndExtract(task)
	if err == nil {
		_, _, currentPath := s.GetPluginPaths(task.Name)
		if task.Name == "PHP" {
			_ = service.UpdatePHPPath(currentPath, true)
		} else if task.Name == "Python" {
			_ = service.UpdatePythonPath(currentPath, true)
		}
		s.Orchestrator.RequestRefresh()
	}
	return err
}

func (s *ServiceManager) InstallPluginModule(parentName string, moduleName string) error {
	_, _, currentPath := s.GetPluginPaths(parentName)
	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("%s is not installed or active", parentName)
	}
	emitProgress := func(name string, pct float64, status string) {
		wruntime.EventsEmit(s.Ctx, "download_progress", plugins.Progress{Name: name, Percentage: pct, Status: status})
	}
	var err error
	switch parentName {
	case "PHP":
		err = php.InstallModule(s.Ctx, s.Downloader, moduleName, currentPath, emitProgress)
		if err == nil { _ = service.UpdatePHPPath(currentPath, true) }
	case "Python":
		err = python.InstallModule(s.Ctx, s.Downloader, moduleName, currentPath, emitProgress)
		if err == nil { _ = service.UpdatePythonPath(currentPath, true) }
	default:
		err = fmt.Errorf("unsupported parent plugin: %s", parentName)
	}
	if err == nil { s.Orchestrator.RequestRefresh() }
	return err
}

func (s *ServiceManager) UninstallPluginModule(parentName string, moduleName string) error {
	_, _, currentPath := s.GetPluginPaths(parentName)
	var err error
	switch parentName {
	case "PHP":
		err = php.UninstallModule(moduleName, currentPath)
		if err == nil { _ = service.UpdatePHPPath(currentPath, true) }
	case "Python":
		err = python.UninstallModule(moduleName, currentPath)
		if err == nil { _ = service.UpdatePythonPath(currentPath, true) }
	default:
		err = fmt.Errorf("unsupported parent plugin: %s", parentName)
	}
	if err == nil { s.Orchestrator.RequestRefresh() }
	return err
}

func (s *ServiceManager) SwitchServiceVersion(serviceName string, version string) error {
	category, binDir, currentPath := s.GetPluginPaths(serviceName)
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
	wasRunning := s.Orchestrator.IsRunning(serviceName)
	if wasRunning {
		_ = s.StopService(serviceName)
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
		return s.StartService(serviceName)
	}
	s.Orchestrator.RequestRefresh()
	return nil
}

func (s *ServiceManager) GetServiceTargetDir(category string, binDir string) string {
	exeMap := map[string]string{
		"nginx":  exeNginx,
		"apache": exeApache,
		"mysql":  exeMySQL,
		"nodejs": exeNode,
		"python": exePython,
	}
	exeName, ok := exeMap[category]
	if !ok { return binDir }
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
