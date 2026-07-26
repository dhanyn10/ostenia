package backend

import (
	"ostenia/internal/config"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// SelectDefaultEditor opens a file dialog to choose the default text editor
func (a *App) SelectDefaultEditor() (string, error) {
	selected, err := a.runtime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Select Default Text Editor",
		Filters: []wruntime.FileFilter{
			{DisplayName: "Executables (*.exe;*.app)", Pattern: "*.exe;*.app"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err == nil && selected != "" {
		a.cfg.DefaultEditor = selected
		_ = config.SaveConfig(a.cfg)
	}
	return selected, err
}

// SetDefaultEditor sets the path to the default text editor manually
func (a *App) SetDefaultEditor(editor string) error {
	a.cfg.DefaultEditor = editor
	return config.SaveConfig(a.cfg)
}

// GetConfig returns the current application configuration
func (a *App) GetConfig() *config.Config {
	return a.cfg
}

// UpdateActiveTab notifies the orchestrator about the current active tab in the UI
func (a *App) UpdateActiveTab(tab string) {
	a.orchestrator.SetActiveTab(tab)
}
