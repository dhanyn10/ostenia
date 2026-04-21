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

func (a *App) InstallPrerequisite(task download.DownloadTask) error {
	fmt.Printf("[App] Installing prerequisite: %s (%s)\n", task.Name, task.URL)
	err := a.downloader.DownloadAndExtract(task)
	if err != nil {
		fmt.Printf("[App] Error installing %s: %v\n", task.Name, err)
	}
	return err
}

func (a *App) CancelDownload(name string) {
	a.downloader.CancelDownload(name)
}
func (a *App) StartService(name string) error {
	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin")

	switch name {
	case "MySQL":
		// Cari mysqld.exe secara dinamis di dalam folder bin/mysql
		var mysqlBin string
		filepath.Walk(filepath.Join(binDir, "mysql"), func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && info.Name() == "mysqld.exe" {
				mysqlBin = path
				return filepath.SkipDir
			}
			return nil
		})

		if mysqlBin == "" {
			return fmt.Errorf("mysqld.exe not found")
		}

		// Cari my.ini di folder yang sama dengan bin/../my.ini
		iniPath := filepath.Join(filepath.Dir(filepath.Dir(mysqlBin)), "my.ini")
		return a.orchestrator.StartService("MySQL", mysqlBin, []string{"--defaults-file=" + iniPath}, filepath.Dir(mysqlBin))

	case "Apache":
		// Cari folder apache secara dinamis
		var apacheBase string
		var apacheBin string
		
		// Priority: folder yang bukan symlink atau symlink yang valid
		filepath.Walk(filepath.Join(binDir, "apache"), func(path string, info os.FileInfo, err error) error {
			if info == nil { return nil }
			
			// Jika kita menemukan httpd.exe
			if !info.IsDir() && info.Name() == "httpd.exe" {
				// Cek apakah path ini valid (bukan bagian dari symlink yang rusak)
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

		// Find available port
		port, err := network.GetAvailablePort([]int{80, 8080, 8081, 9000})
		if err != nil {
			return fmt.Errorf("no available ports for apache: %v", err)
		}

		// Set dynamic PATH
		phpPath := filepath.Join(baseDir, "bin", "php", "current")
		nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")
		os.Setenv("PATH", os.Getenv("PATH")+";"+phpPath+";"+nodePath)
		
		a.updateApacheConfig(apacheBase, port)

		// Start Apache using httpd.exe directly for better control
		return a.orchestrator.StartService("Apache", apacheBin, []string{}, apacheBase)
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
	}
func (a *App) StopService(name string) {
	a.orchestrator.StopService(name)
}

func (a *App) StartAllServices() error {
	a.StartService("MySQL")
	return a.StartService("Apache")
}

func (a *App) updateApacheConfig(apachePath string, port int) {
	baseDir := config.GetBaseDir()
	wwwDir := a.cfg.WWWRoot

	// Ensure WWWRoot exists
	os.MkdirAll(wwwDir, 0755)

	files, _ := os.ReadDir(wwwDir)
	var vhostsContent string
	for _, f := range files {
		if f.IsDir() {
			projectName := f.Name()
			projectPath := filepath.Join(wwwDir, projectName)
			vhostsContent += service.GenerateVHost(projectName, projectPath)

			// Add to hosts file (requires admin usually, but we'll try)
			network.AddHost("127.0.0.1", projectName+".test")

			// Sign SSL cert for project
			caDir := filepath.Join(baseDir, "ssl")
			certDir := filepath.Join(projectPath, ".ssl")
			os.MkdirAll(certDir, 0755)
			ssl.SignCertificate(caDir, projectName+".test", certDir)
		}
	}

	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	
	// Detect PHP DLL for Apache
	var phpDll string
	phpEntries, _ := os.ReadDir(phpPath)
	for _, entry := range phpEntries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "php") && strings.Contains(entry.Name(), "apache2_4") && strings.HasSuffix(entry.Name(), ".dll") {
			phpDll = filepath.Join(phpPath, entry.Name())
			break
		}
	}
	if phpDll == "" {
		phpDll = filepath.Join(phpPath, "php8apache2_4.dll")
	}

	service.UpdateApacheConfig(apachePath, phpDll, phpPath, vhostsContent, port)
}

func (a *App) StopAllServices() {
	a.orchestrator.StopAll()
}

func (a *App) OpenTerminal() {
	baseDir := config.GetBaseDir()
	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	mysqlPath := filepath.Join(baseDir, "bin", "mysql", "current", "bin")
	nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")

	// Prepare environment
	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if len(e) >= 5 && e[:5] == "PATH=" {
			env[i] = "PATH=" + phpPath + ";" + mysqlPath + ";" + nodePath + ";" + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+phpPath+";"+mysqlPath+";"+nodePath)
	}

	// Laragon style terminal (CMD)
	// We'll try to start it in the WWWRoot
	cmd := service.NewTerminal(a.cfg.WWWRoot, env)
	cmd.Start()
}

func (a *App) DeleteVersion(taskName, version string) error {
	return a.downloader.DeleteVersion(taskName, version)
}
