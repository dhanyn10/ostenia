package backend

// ToggleDevTools toggles the browser developer tools
func (a *App) ToggleDevTools() {
	a.runtime.WindowExecJS(a.ctx, "window.runtime.WindowToggleDevTools()")
}

// Minimize minimizes the application window
func (a *App) Minimize() { a.runtime.WindowMinimise(a.ctx) }

// Maximize maximizes the application window
func (a *App) Maximize() { a.runtime.WindowMaximise(a.ctx) }

// Unmaximize restores the application window from maximized state
func (a *App) Unmaximize() { a.runtime.WindowUnmaximise(a.ctx) }

// Close closes the application
func (a *App) Close() { a.runtime.Quit(a.ctx) }
