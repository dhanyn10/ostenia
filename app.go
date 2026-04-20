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

func (a *App) CancelDownload(name string) {
	a.downloader.CancelDownload(name)
}

func (a *App) StartService(name string) error {
	baseDir := config.GetBaseDir()
	switch name {
	case "MySQL":
		mysqlBin := filepath.Join(baseDir, "bin", "mysql", "current", "bin", "mysqld.exe")
		return a.orchestrator.StartService("MySQL", mysqlBin, []string{"--defaults-file=" + filepath.Join(baseDir, "bin", "mysql", "current", "my.ini")})
	case "Apache":
		apacheBin := filepath.Join(baseDir, "bin", "apache", "current", "bin", "httpd.exe")
		// Set dynamic PATH
		phpPath := filepath.Join(baseDir, "bin", "php", "current")
		nodePath := filepath.Join(baseDir, "bin", "nodejs", "current")
		os.Setenv("PATH", os.Getenv("PATH")+";"+phpPath+";"+nodePath)
		a.updateVHosts()
		return a.orchestrator.StartService("Apache", apacheBin, []string{})
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
