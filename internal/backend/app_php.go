package backend

import (
	"context"
	"ostenia/internal/config"
	"ostenia/internal/service"
	"path/filepath"
	"time"
)

// GetPHPExtensions returns the list of PHP extensions and their status from php.ini
func (a *App) GetPHPExtensions() ([]service.PHPExtensionInfo, error) {
	baseDir := config.GetBaseDir()
	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	return service.GetPHPExtensions(phpPath)
}

// TogglePHPExtension enables or disables a PHP extension in php.ini
func (a *App) TogglePHPExtension(ctx context.Context, extName string, enable bool) error {
	baseDir := config.GetBaseDir()
	phpPath := filepath.Join(baseDir, "bin", "php", "current")
	err := service.TogglePHPExtension(phpPath, extName, enable)
	if err != nil {
		return err
	}
	if a.orchestrator.IsRunning("PHP") {
		_ = a.StopService(ctx, "PHP")
		time.Sleep(600 * time.Millisecond)
		return a.StartService(ctx, "PHP")
	}
	return nil
}
