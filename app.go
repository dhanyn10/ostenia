package main

import (
	"context"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/download"
	"ostenia/internal/network"
	"ostenia/internal/service"
	"ostenia/internal/ssl"
	"path/filepath"
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
	return a.downloader.DownloadAndExtract(task)
}

func (a *App) StartAllServices() error {
	baseDir := config.GetBaseDir()

	// 1. Start MySQL
	mysqlBin := filepath.Join(baseDir, "bin", "mysql", "current", "bin", "mysqld.exe")
	a.orchestrator.StartService("MySQL", mysqlBin, []string{"--defaults-file=" + filepath.Join(baseDir, "bin", "mysql", "current", "my.ini")})

	// 2. Start Apache
	apacheBin := filepath.Join(baseDir, "bin", "apache", "current", "bin", "httpd.exe")

	// Set dynamic PATH so Apache/PHP can find each other and Node.js
	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")
	os.Setenv("PATH", os.Getenv("PATH")+";"+phpPath+";"+nodePath)

	// Update Apache Config with VHosts
	a.updateVHosts()

	return a.orchestrator.StartService("Apache", apacheBin, []string{})
}

func (a *App) updateVHosts() {
	baseDir := config.GetBaseDir()
	wwwDir := a.cfg.WWWRoot

	files, _ := os.ReadDir(wwwDir)
	var vhostsContent string
	for _, f := range files {
		if f.IsDir() {
			projectName := f.Name()
			projectPath := filepath.Join(wwwDir, projectName)
			vhostsContent += service.GenerateVHost(projectName, projectPath)

			// Add to hosts file
			network.AddHost("127.0.0.1", projectName+".test")

			// Sign SSL cert for project
			caDir := filepath.Join(baseDir, "ssl")
			certDir := filepath.Join(projectPath, ".ssl")
			os.MkdirAll(certDir, 0755)
			ssl.SignCertificate(caDir, projectName+".test", certDir)
		}
	}

	apachePath := filepath.Join(baseDir, "bin", "apache", "current")
	phpDll := filepath.Join(baseDir, "bin", "php", "current", "php8apache2_4.dll")
	service.UpdateApacheConfig(apachePath, phpDll, vhostsContent)
}

func (a *App) StopAllServices() {
	a.orchestrator.StopAll()
}
