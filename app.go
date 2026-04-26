package main

import (
	"context"
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/download"
	"ostenia/internal/network"
	"ostenia/internal/service"
	"ostenia/internal/ssl"
	"path/filepath"
	"strings"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	downloader   *download.Manager
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
	a.downloader = download.NewManager(ctx)
	a.orchestrator = service.NewOrchestrator(ctx)
	a.symlinkMgr = service.NewSymlinkManager()

	cfg, _ := config.LoadConfig()
	a.cfg = cfg

	// Ensure SSL directory exists
	baseDir := config.GetBaseDir()
	caDir := filepath.Join(baseDir, "ssl")
	os.MkdirAll(caDir, 0755)
}

func (a *App) GetPrerequisites() []download.DownloadTask {
	return download.GetLatestKnownVersions()
}

func (a *App) GetServiceStatus(serviceName string) service.ServiceDetailedInfo {
	return a.orchestrator.GetDetailedInfo(serviceName)
}

func (a *App) GetServerRoot() string {
	if a.cfg == nil {
		return ""
	}
	return a.cfg.WWWRoot
}

func (a *App) SetServerRoot(rootPath string) error {
	fmt.Printf("[App] Setting Server Root to: %s\n", rootPath)
	a.cfg.WWWRoot = rootPath
	err := config.SaveConfig(a.cfg)
	if err != nil {
		return err
	}

	// Restart active web servers
	if a.orchestrator.IsRunning("Apache") {
		a.StopService("Apache")
		a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		a.StopService("Nginx")
		a.StartService("Nginx")
	}

	return nil
}

func (a *App) SelectServerRoot() (string, error) {
	selectedDir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Select Server Root Directory",
	})
	if err != nil {
		return "", err
	}
	if selectedDir != "" {
		err = a.SetServerRoot(selectedDir)
		if err != nil {
			return "", err
		}
	}
	return selectedDir, nil
}

func (a *App) OpenServerRootFolder() error {
	if a.cfg == nil || a.cfg.WWWRoot == "" {
		return fmt.Errorf("server root directory not set")
	}
	return service.OpenExplorer(a.cfg.WWWRoot)
}

func (a *App) OpenPluginFolder(serviceName string) error {
	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin")

	// Determine category folder
	category := strings.ToLower(serviceName)
	folderPath := filepath.Join(binDir, category)

	// Check if the directory exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return fmt.Errorf("folder for %s not found: %s", serviceName, folderPath)
	}

	// Open folder in explorer
	return service.OpenExplorer(folderPath)
}

func (a *App) InstallPrerequisite(task download.DownloadTask) error {
	fmt.Printf("[App] Installing prerequisite: %s (%s)\n", task.Name, task.URL)
	err := a.downloader.DownloadAndExtract(task)
	if err != nil {
		fmt.Printf("[App] Error installing %s: %v\n", task.Name, err)
	}
	return err
}

func (a *App) CancelDownload(taskName string) {
	a.downloader.CancelDownload(taskName)
}

func (a *App) StartService(serviceName string) error {
	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin")

	fmt.Printf("[App] Starting service: %s\n", serviceName)

	switch serviceName {
	case "OpenSSL":
		caDir := filepath.Join(baseDir, "ssl")
		err := ssl.GenerateRootCA(caDir)
		if err != nil {
			return err
		}
		// Refresh status manually as OpenSSL isn't a long-running process
		info := a.orchestrator.GetDetailedInfo("OpenSSL")
		wruntime.EventsEmit(a.ctx, "service_status", info)
		return nil

	case "MySQL":
		var mysqlBin string
		var mysqlBase string
		filepath.Walk(filepath.Join(binDir, "mysql"), func(path string, info os.FileInfo, err error) error {
			if info == nil { return nil }
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

		port, err := network.GetAvailablePort([]int{3306, 3307, 3308})
		if err != nil {
			return fmt.Errorf("no available ports for mysql: %v", err)
		}

		os.Setenv("PATH", filepath.Dir(mysqlBin)+";"+os.Getenv("PATH"))
		err = a.updateMySQLConfig(mysqlBase, port)
		if err != nil { return err }

		iniPath := filepath.Join(mysqlBase, "my.ini")
		return a.orchestrator.StartServiceWithPort("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath, "--console"}, filepath.Dir(mysqlBin), port)

	case "Apache":
		if a.orchestrator.IsRunning("Nginx") {
			a.StopService("Nginx")
		}

		var apacheBase string
		var apacheBin string
		filepath.Walk(filepath.Join(binDir, "apache"), func(path string, info os.FileInfo, err error) error {
			if info == nil { return nil }
			if !info.IsDir() && info.Name() == "httpd.exe" {
				if _, err := os.Stat(path); err == nil {
					apacheBin = path
					apacheBase = filepath.Dir(filepath.Dir(path))
					return filepath.SkipDir
				}
			}
			return nil
		})

		if apacheBase == "" { return fmt.Errorf("apache installation not found") }

		port, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil { return fmt.Errorf("no available ports for apache: %v", err) }

		phpPath := filepath.Join(baseDir, "bin", "php", "current")
		nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")
		os.Setenv("PATH", phpPath+";"+os.Getenv("PATH")+";"+nodePath)
		
		err = a.updateApacheConfig(apacheBase, port)
		if err != nil { return err }

		return a.orchestrator.StartServiceWithPort("Apache", apacheBin, []string{}, apacheBase, port)

	case "Nginx":
		if a.orchestrator.IsRunning("Apache") {
			a.StopService("Apache")
		}

		var nginxBase string
		var nginxBin string
		filepath.Walk(filepath.Join(binDir, "nginx"), func(path string, info os.FileInfo, err error) error {
			if info == nil { return nil }
			if !info.IsDir() && info.Name() == "nginx.exe" {
				nginxBin = path
				nginxBase = filepath.Dir(path)
				return filepath.SkipDir
			}
			return nil
		})

		if nginxBase == "" { return fmt.Errorf("nginx installation not found") }

		port, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil { return fmt.Errorf("no available ports for nginx: %v", err) }

		err = a.updateNginxConfig(nginxBase, port)
		if err != nil { return err }

		return a.orchestrator.StartServiceWithPort("Nginx", nginxBin, []string{"-p", nginxBase}, nginxBase, port)

	case "HeidiSQL":
		heidisqlBin := filepath.Join(binDir, "heidisql", "heidisql.exe")
		if _, err := os.Stat(heidisqlBin); os.IsNotExist(err) { return fmt.Errorf("heidisql.exe not found") }
		return a.orchestrator.StartService("HeidiSQL", heidisqlBin, []string{}, filepath.Dir(heidisqlBin))

	case "PHP":
		phpPath := filepath.Join(baseDir, "bin", "php", "current")
		phpCgi := filepath.Join(phpPath, "php-cgi.exe")
		if _, err := os.Stat(phpCgi); os.IsNotExist(err) { return fmt.Errorf("php-cgi.exe not found") }

		port, err := network.GetAvailablePort([]int{9000, 9001, 9002})
		if err != nil { return fmt.Errorf("no available ports for PHP: %v", err) }

		os.Setenv("PHP_FCGI_MAX_REQUESTS", "1000")
		err = a.orchestrator.StartServiceWithPort("PHP", phpCgi, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, phpPath, port)

		if err == nil {
			if a.orchestrator.IsRunning("Apache") { a.StopService("Apache"); a.StartService("Apache") }
			if a.orchestrator.IsRunning("Nginx") { a.StopService("Nginx"); a.StartService("Nginx") }
		}
		return err

	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

func (a *App) StopService(serviceName string) {
	if serviceName == "OpenSSL" {
		baseDir := config.GetBaseDir()
		caDir := filepath.Join(baseDir, "ssl")
		os.RemoveAll(caDir)
		os.MkdirAll(caDir, 0755)

		info := a.orchestrator.GetDetailedInfo("OpenSSL")
		wruntime.EventsEmit(a.ctx, "service_status", info)
		return
	}
	a.orchestrator.StopService(serviceName)
}

func (a *App) StartAllServices() error {
	a.StartService("MySQL")
	a.StartService("PHP")
	return a.StartService("Apache")
}

func (a *App) updateMySQLConfig(mysqlPath string, port int) error {
	dataDir := filepath.Join(mysqlPath, "data")
	tmpDir := filepath.Join(mysqlPath, "tmp")
	binDir := filepath.Join(mysqlPath, "bin")
	iniPath := filepath.Join(mysqlPath, "my.ini")
	err := service.UpdateMySQLConfig(mysqlPath, dataDir, tmpDir, port)
	if err != nil { return err }
	return service.InitializeMySQLDataDir(binDir, mysqlPath, dataDir, iniPath)
}

func (a *App) updateApacheConfig(apachePath string, port int) error {
	if port <= 0 { port = 80 }
	wwwDir := a.cfg.WWWRoot
	os.MkdirAll(wwwDir, 0755)

	phpInfo := a.orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" { phpPort = phpInfo.Port }

	return service.UpdateApacheConfig(apachePath, "", "", "", port, wwwDir, phpPort)
}

func (a *App) updateNginxConfig(nginxPath string, port int) error {
	if port <= 0 { port = 80 }
	wwwDir := a.cfg.WWWRoot
	os.MkdirAll(wwwDir, 0755)

	phpInfo := a.orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" { phpPort = phpInfo.Port }

	return service.UpdateNginxConfig(nginxPath, wwwDir, phpPort, port)
}

func (a *App) StopAllServices() {
	a.orchestrator.StopAll()
}

func (a *App) OpenTerminal(terminalType string) {
	baseDir := config.GetBaseDir()
	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	mysqlPath := filepath.Join(baseDir, "bin", "mysql", "current", "bin")
	nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")

	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = "PATH=" + phpPath + ";" + mysqlPath + ";" + nodePath + ";" + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound { env = append(env, "PATH="+phpPath+";"+mysqlPath+";"+nodePath) }

	cmd := service.NewTerminal(a.cfg.WWWRoot, env)
	cmd.Open(terminalType)
}

func (a *App) DeleteVersion(serviceName string, version string) error {
	return a.downloader.DeleteVersion(serviceName, version)
}
