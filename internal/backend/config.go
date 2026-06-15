package backend

import (
	"ostenia/internal/config"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"context"
)

type ConfigManager struct {
	Ctx    context.Context
	Config *config.Config
}

func (c *ConfigManager) GetConfig() *config.Config {
	return c.Config
}

func (c *ConfigManager) SaveConfig(cfg *config.Config) error {
	return config.SaveConfig(cfg)
}

func (c *ConfigManager) SetDefaultEditor(editorPath string) error {
	c.Config.DefaultEditor = editorPath
	return config.SaveConfig(c.Config)
}

func (c *ConfigManager) SelectDefaultEditor() (string, error) {
	selected, err := wruntime.OpenFileDialog(c.Ctx, wruntime.OpenDialogOptions{
		Title: "Select Default Editor Executable",
		Filters: []wruntime.FileFilter{{DisplayName: "Executables", Pattern: "*.exe;*.app;*"}},
	})
	if err != nil { return "", err }
	if selected != "" {
		c.Config.DefaultEditor = selected
		err = config.SaveConfig(c.Config)
	}
	return selected, err
}
