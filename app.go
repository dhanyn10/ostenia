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

	// Ensure Root CA exists
	baseDir := config.GetBaseDir()
	caDir := filepath.Join(baseDir, "ssl")
	os.MkdirAll(caDir, 0755)
	ssl.GenerateRootCA(caDir)
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

	if a.orchestrator.IsRunning("Apache") {
		fmt.Println("[App] Apache is running, restarting...")
		a.StopService("Apache")
		return a.StartService("Apache")
	}

	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin")
	var apacheBase string
	filepath.Walk(filepath.Join(binDir, "apache"), func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == "httpd.exe" {
			apacheBase = filepath.Dir(filepath.Dir(path))
			return filepath.SkipDir
		}
		return nil
	})

	if apacheBase != "" {
		fmt.Printf("[App] Updating Apache config at: %s\n", apacheBase)
		return a.updateApacheConfig(apacheBase, 80)
	}

	return nil
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

		fmt.Printf("[App] MySQL port: %d, Base: %s\n", port, mysqlBase)

		err = a.updateMySQLConfig(mysqlBase, port)
		if err != nil {
			fmt.Printf("[App] Error updating MySQL config: %v\n", err)
			return err
		}

		iniPath := filepath.Join(mysqlBase, "my.ini")
		return a.orchestrator.StartServiceWithPort("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath, "--console"}, filepath.Dir(mysqlBin), port)

	case "Apache":
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

		if apacheBase == "" {
			return fmt.Errorf("apache installation not found")
		}

		port, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil {
			return fmt.Errorf("no available ports for apache: %v", err)
		}

		phpPath := filepath.Join(baseDir, "bin", "php", "current")
		nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")
		os.Setenv("PATH", phpPath+";"+os.Getenv("PATH")+";"+nodePath)
		
		err = a.updateApacheConfig(apacheBase, port)
		if err != nil {
			return err
		}

		return a.orchestrator.StartServiceWithPort("Apache", apacheBin, []string{}, apacheBase, port)

	case "HeidiSQL":
		heidisqlBin := filepath.Join(binDir, "heidisql", "heidisql.exe")
		if _, err := os.Stat(heidisqlBin); os.IsNotExist(err) {
			return fmt.Errorf("heidisql.exe not found")
		}
		return a.orchestrator.StartService("HeidiSQL", heidisqlBin, []string{}, filepath.Dir(heidisqlBin))

	case "PHP":
		phpPath := filepath.Join(baseDir, "bin", "php", "current")
		phpCgi := filepath.Join(phpPath, "php-cgi.exe")
		if _, err := os.Stat(phpCgi); os.IsNotExist(err) {
			return fmt.Errorf("php-cgi.exe not found")
		}

		port, err := network.GetAvailablePort([]int{9000, 9001, 9002})
		if err != nil {
			return fmt.Errorf("no available ports for PHP: %v", err)
		}

		os.Setenv("PHP_FCGI_MAX_REQUESTS", "1000")

		err = a.orchestrator.StartServiceWithPort("PHP", phpCgi, []string{"-b", fmt.Sprintf("127.0.0.1:%d", port)}, phpPath, port)

		if err == nil && a.orchestrator.IsRunning("Apache") {
			var apacheBase string
			filepath.Walk(filepath.Join(binDir, "apache"), func(path string, info os.FileInfo, err error) error {
				if info != nil && !info.IsDir() && info.Name() == "httpd.exe" {
					apacheBase = filepath.Dir(filepath.Dir(path))
					return filepath.SkipDir
				}
				return nil
			})
			if apacheBase != "" {
				a.updateApacheConfig(apacheBase, 0)
			}
		}
		return err

	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

func (a *App) StopService(serviceName string) {
	a.orchestrator.StopService(serviceName)

	if serviceName == "Apache" || serviceName == "PHP" {
		baseDir := config.GetBaseDir()
		binDir := filepath.Join(baseDir, "bin")
		var apacheBase string
		filepath.Walk(filepath.Join(binDir, "apache"), func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == "httpd.exe" {
				apacheBase = filepath.Dir(filepath.Dir(path))
				return filepath.SkipDir
			}
			return nil
		})
		if apacheBase != "" {
			a.updateApacheConfig(apacheBase, 80)
		}
	}
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

	fmt.Printf("[App] Updating MySQL Config at: %s\n", mysqlPath)

	err := service.UpdateMySQLConfig(mysqlPath, dataDir, tmpDir, port)
	if err != nil {
		return err
	}

	return service.InitializeMySQLDataDir(binDir, mysqlPath, dataDir, iniPath)
}

func (a *App) updateApacheConfig(apachePath string, port int) error {
	baseDir := config.GetBaseDir()
	wwwDir := a.cfg.WWWRoot

	os.MkdirAll(wwwDir, 0755)

	files, _ := os.ReadDir(wwwDir)
	var vhostsContent string
	for _, f := range files {
		if f.IsDir() {
			projectName := f.Name()
			projectPath := filepath.Join(wwwDir, projectName)
			vhostsContent += service.GenerateVHost(projectName, projectPath, port)
			network.AddHost("127.0.0.1", projectName+".test")
			caDir := filepath.Join(baseDir, "ssl")
			certDir := filepath.Join(projectPath, ".ssl")
			os.MkdirAll(certDir, 0755)
			ssl.SignCertificate(caDir, projectName+".test", certDir)
		}
	}

	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	
	var phpDll string
	if entries, err := os.ReadDir(phpPath); err == nil {
		for _, entry := range entries {
			name := strings.ToLower(entry.Name())
			if !entry.IsDir() && strings.HasPrefix(name, "php") && strings.Contains(name, "apache2_4") && strings.HasSuffix(name, ".dll") {
				phpDll = filepath.Join(phpPath, entry.Name())
				break
			}
		}
	}

	if phpDll == "" {
		potentialDll := filepath.Join(phpPath, "php8apache2_4.dll")
		if _, err := os.Stat(potentialDll); err == nil {
			phpDll = potentialDll
		}
	}

	phpInfo := a.orchestrator.GetDetailedInfo("PHP")
	phpPort := 0
	if phpInfo.Status == "Running" {
		phpPort = phpInfo.Port
	}

	return service.UpdateApacheConfig(apachePath, phpDll, phpPath, vhostsContent, port, wwwDir, phpPort)
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
	if !pathFound {
		env = append(env, "PATH="+phpPath+";"+mysqlPath+";"+nodePath)
	}

	cmd := service.NewTerminal(a.cfg.WWWRoot, env)
	cmd.Open(terminalType)
}

func (a *App) DeleteVersion(serviceName string, version string) error {
	return a.downloader.DeleteVersion(serviceName, version)
}
