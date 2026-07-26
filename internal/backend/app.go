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

// App struct manages the main application state and coordinates between backend services and the frontend.
// It wraps core controllers including plugins downloaders, orchestrators, symlinks, SSH, and SSL managers.
type App struct {
	ctx          context.Context             // Main application startup context (stores Wails session state)
	downloader   interfaces.PluginManager    // Controller responsible for plugin downloads and extraction
	orchestrator interfaces.Orchestrator     // Controller managing lifecycle, watching, and state of development services
	symlinkMgr   *service.SymlinkManager     // Helper maintaining system paths and junction points
	sshManager   interfaces.SSHManager       // Controller executing SSH/WSL connection streams
	sslManager   interfaces.SSLManager       // Controller orchestrating local Root CAs and signed certs
	cfg          *config.Config              // Global persistent configuration settings of Ostenia
	runtime      interfaces.Runtime          // Wrapper to trigger Wails window, dialog, and event operations
}

// WailsRuntime implements interfaces.Runtime to forward window and event actions directly to the standard Wails runtime.
type WailsRuntime struct{}

// EventsEmit broadcasts an event notification to the frontend.
func (w *WailsRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
	wruntime.EventsEmit(ctx, eventName, optionalData...)
}

// WindowMinimise minimizes the Wails application window.
func (w *WailsRuntime) WindowMinimise(ctx context.Context) { wruntime.WindowMinimise(ctx) }

// WindowMaximise maximizes the Wails application window.
func (w *WailsRuntime) WindowMaximise(ctx context.Context) { wruntime.WindowMaximise(ctx) }

// WindowUnmaximise unmaximizes (restores) the Wails application window.
func (w *WailsRuntime) WindowUnmaximise(ctx context.Context) { wruntime.WindowUnmaximise(ctx) }

// WindowExecJS executes custom JavaScript in the frontend browser context.
func (w *WailsRuntime) WindowExecJS(ctx context.Context, js string) { wruntime.WindowExecJS(ctx, js) }

// Quit gracefully closes the Wails application process.
func (w *WailsRuntime) Quit(ctx context.Context) { wruntime.Quit(ctx) }

// OpenFileDialog shows a file selection dialog on the host screen.
func (w *WailsRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return wruntime.OpenFileDialog(ctx, options)
}

// OpenDirectoryDialog shows a directory selection dialog on the host screen.
func (w *WailsRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return wruntime.OpenDirectoryDialog(ctx, options)
}

// SaveFileDialog shows a file save dialog on the host screen.
func (w *WailsRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return wruntime.SaveFileDialog(ctx, options)
}

// EventsEmit triggers a standard event emission to the frontend browser using App context.
func (a *App) EventsEmit(eventName string, optionalData ...interface{}) {
	a.runtime.EventsEmit(a.ctx, eventName, optionalData...)
}

// Quit initiates graceful application exit.
func (a *App) Quit() { a.runtime.Quit(a.ctx) }

// OpenFileDialog opens standard file picker views on top of the browser view.
func (a *App) OpenFileDialog(options wruntime.OpenDialogOptions) (string, error) {
	return a.runtime.OpenFileDialog(a.ctx, options)
}

// OpenDirectoryDialog opens directory picker views on top of the browser view.
func (a *App) OpenDirectoryDialog(options wruntime.OpenDialogOptions) (string, error) {
	return a.runtime.OpenDirectoryDialog(a.ctx, options)
}

// SaveFileDialog opens file saver prompts on top of the browser view.
func (a *App) SaveFileDialog(options wruntime.SaveDialogOptions) (string, error) {
	return a.runtime.SaveFileDialog(a.ctx, options)
}

// GenerateRootCA creates a new local SSL Root Certificate Authority keypair.
func (a *App) GenerateRootCA(destDir string) error { return a.sslManager.GenerateRootCA(destDir) }

// GetRemainingDays calculates the valid lifetime span left for a targeted SSL certificate.
func (a *App) GetRemainingDays(certPath string) (int, error) {
	return a.sslManager.GetRemainingDays(certPath)
}

// SignCertificate signs a custom domain domain-level SSL certificate using the generated local Root CA.
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

// NewApp creates and initializes a clean App struct instance.
func NewApp() *App {
	return &App{}
}

// DefaultSSLManager provides default production SSL wrapping endpoints.
type DefaultSSLManager struct{}

// GenerateRootCA generates a default local authority certificate pair.
func (s *DefaultSSLManager) GenerateRootCA(destDir string) error { return ssl.GenerateRootCA(destDir) }

// GetRemainingDays retrieves remaining validity days for a local certificate file.
func (s *DefaultSSLManager) GetRemainingDays(path string) (int, error) {
	return ssl.GetRemainingDays(path)
}

// SignCertificate issues a local domain-specific SSL certificate.
func (s *DefaultSSLManager) SignCertificate(caDir, domain, destDir string) error {
	return ssl.SignCertificate(caDir, domain, destDir)
}

// Startup is called when the application initializes.
// Stores the application context and instantiates standard production dependencies.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
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

	// Setup necessary development folders inside the root execution directory
	a.ensureEnvironmentStructure()

	// Spin up service monitors in the background
	a.orchestrator.StartWatcher(ctx)

	// Watch and map proxy ports
	go a.startProxyWatcher(ctx)
}
