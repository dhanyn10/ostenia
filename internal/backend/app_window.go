package backend

import "context"

// ToggleDevTools toggles the browser developer tools
func (a *App) ToggleDevTools(ctx context.Context) {
	if ctx != nil {
		a.runtime.WindowExecJS(ctx, "window.runtime.WindowToggleDevTools()")
	}
}

// Minimize minimizes the application window
func (a *App) Minimize(ctx context.Context) {
	if ctx != nil {
		a.runtime.WindowMinimise(ctx)
	}
}

// Maximize maximizes the application window
func (a *App) Maximize(ctx context.Context) {
	if ctx != nil {
		a.runtime.WindowMaximise(ctx)
	}
}

// Unmaximize restores the application window from maximized state
func (a *App) Unmaximize(ctx context.Context) {
	if ctx != nil {
		a.runtime.WindowUnmaximise(ctx)
	}
}

// Close closes the application
func (a *App) Close(ctx context.Context) {
	if ctx != nil {
		a.runtime.Quit(ctx)
	}
}
