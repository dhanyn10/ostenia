package backend

import wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

// ToggleDevTools toggles the browser developer tools
func (a *App) ToggleDevTools() {
	wruntime.WindowExecJS(a.ctx, "window.runtime.WindowToggleDevTools()")
}

// Minimize minimizes the application window
func (a *App) Minimize() { wruntime.WindowMinimise(a.ctx) }

// Maximize maximizes the application window
func (a *App) Maximize() { wruntime.WindowMaximise(a.ctx) }

// Unmaximize restores the application window from maximized state
func (a *App) Unmaximize() { wruntime.WindowUnmaximise(a.ctx) }

// Close closes the application
func (a *App) Close() { wruntime.Quit(a.ctx) }
