package backend

import (
	"context"
	"fmt"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/service"
	"ostenia/internal/ssl"
)

// App struct manages the main application state and coordinates between backend services and the frontend
type App struct {
	downloader   interfaces.PluginManager
	orchestrator interfaces.Orchestrator
	symlinkMgr   *service.SymlinkManager
	sshManager   interfaces.SSHManager
	sslManager   interfaces.SSLManager
	cfg          *config.Config
	runtime      interfaces.Runtime
}

type WailsRuntime struct{}

func (w *WailsRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
	wruntime.EventsEmit(ctx, eventName, optionalData...)
}
func (w *WailsRuntime) WindowMinimise(ctx context.Context)          { wruntime.WindowMinimise(ctx) }
func (w *WailsRuntime) WindowMaximise(ctx context.Context)          { wruntime.WindowMaximise(ctx) }
func (w *WailsRuntime) WindowUnmaximise(ctx context.Context)        { wruntime.WindowUnmaximise(ctx) }
func (w *WailsRuntime) WindowExecJS(ctx context.Context, js string) { wruntime.WindowExecJS(ctx, js) }
func (w *WailsRuntime) Quit(ctx context.Context)                    { wruntime.Quit(ctx) }
func (w *WailsRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return wruntime.OpenFileDialog(ctx, options)
}
func (w *WailsRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return wruntime.OpenDirectoryDialog(ctx, options)
}
func (w *WailsRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return wruntime.SaveFileDialog(ctx, options)
}

func (a *App) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
	a.runtime.EventsEmit(ctx, eventName, optionalData...)
}
func (a *App) Quit(ctx context.Context) { a.runtime.Quit(ctx) }
func (a *App) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return a.runtime.OpenFileDialog(ctx, options)
}
func (a *App) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return a.runtime.OpenDirectoryDialog(ctx, options)
}
func (a *App) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return a.runtime.SaveFileDialog(ctx, options)
}

func (a *App) GenerateRootCA(destDir string) error { return a.sslManager.GenerateRootCA(destDir) }
func (a *App) GetRemainingDays(certPath string) (int, error) {
	return a.sslManager.GetRemainingDays(certPath)
}
func (a *App) SignCertificate(caDir string, domain string, destDir string) error {
	return a.sslManager.SignCertificate(caDir, domain, destDir)
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
type DefaultSSLManager struct{}

func (s *DefaultSSLManager) GenerateRootCA(destDir string) error { return ssl.GenerateRootCA(destDir) }
func (s *DefaultSSLManager) GetRemainingDays(path string) (int, error) {
	return ssl.GetRemainingDays(path)
}
func (s *DefaultSSLManager) SignCertificate(caDir, domain, destDir string) error {
	return ssl.SignCertificate(caDir, domain, destDir)
}

func (a *App) Startup(ctx context.Context) {
	if a.runtime == nil {
		a.runtime = &WailsRuntime{}
	}
	if a.downloader == nil {
		a.downloader = plugins.NewManager(ctx)
	}
	if a.orchestrator == nil {
		a.orchestrator = service.NewOrchestrator()
	}
	if a.symlinkMgr == nil {
		a.symlinkMgr = service.NewSymlinkManager()
	}
	if a.sshManager == nil {
		a.sshManager = service.NewSSHManager()
	}
	if a.sslManager == nil {
		a.sslManager = &DefaultSSLManager{}
	}

	a.orchestrator.SetRuntime(a.runtime)
	a.sshManager.SetRuntime(a.runtime)

	cfg, err := config.LoadConfig()
	if err != nil {
		// Log error if config fails to load
		fmt.Printf("[App] Error loading config: %v\n", err)
	}
	a.cfg = cfg

	// Initial setup of directories in current base dir
	a.ensureEnvironmentStructure()

	// Start the periodic watcher for services
	a.orchestrator.StartWatcher(ctx)

	// Start proxy port watcher
	go a.startProxyWatcher(ctx)
}
