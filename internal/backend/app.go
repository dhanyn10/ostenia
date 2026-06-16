package backend

import (
	"context"
	"fmt"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/service"
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

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
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
