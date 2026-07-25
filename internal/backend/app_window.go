package backend

// ToggleDevTools toggles the browser developer tools
func (a *App) ToggleDevTools() {
	if a.ctx != nil {
		a.runtime.WindowExecJS(a.ctx, "window.runtime.WindowToggleDevTools()")
	}
}

// Minimize minimizes the application window
func (a *App) Minimize() {
	if a.ctx != nil {
		a.runtime.WindowMinimise(a.ctx)
	}
}

// Maximize maximizes the application window
func (a *App) Maximize() {
	if a.ctx != nil {
		a.runtime.WindowMaximise(a.ctx)
	}
}

// Unmaximize restores the application window from maximized state
func (a *App) Unmaximize() {
	if a.ctx != nil {
		a.runtime.WindowUnmaximise(a.ctx)
	}
}

// Close closes the application
func (a *App) Close() {
	if a.ctx != nil {
		a.runtime.Quit(a.ctx)
	}
}
