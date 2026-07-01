package backend

import (
	"context"
	"fmt"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/service"
	"ostenia/internal/backend/interfaces"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct manages the main application state and coordinates between backend services and the frontend
type App struct {
	ctx          context.Context
	downloader   *plugins.Manager
	orchestrator *service.Orchestrator
	symlinkMgr   *service.SymlinkManager
	sshManager   *service.SSHManager
	cfg          *config.Config
	runtime      interfaces.Runtime
}

type WailsRuntime struct{}

func (w *WailsRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
	wruntime.EventsEmit(ctx, eventName, optionalData...)
}
func (w *WailsRuntime) WindowMinimise(ctx context.Context) { wruntime.WindowMinimise(ctx) }
func (w *WailsRuntime) WindowMaximise(ctx context.Context) { wruntime.WindowMaximise(ctx) }
func (w *WailsRuntime) WindowUnmaximise(ctx context.Context) { wruntime.WindowUnmaximise(ctx) }
func (w *WailsRuntime) WindowExecJS(ctx context.Context, js string) { wruntime.WindowExecJS(ctx, js) }
func (w *WailsRuntime) Quit(ctx context.Context) { wruntime.Quit(ctx) }
func (w *WailsRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return wruntime.OpenFileDialog(ctx, options)
}
func (w *WailsRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return wruntime.OpenDirectoryDialog(ctx, options)
}
func (w *WailsRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return wruntime.SaveFileDialog(ctx, options)
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
	a.runtime = &WailsRuntime{}
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
