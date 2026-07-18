package service

import (
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
)

type SymlinkManager struct{}

func NewSymlinkManager() *SymlinkManager {
	return &SymlinkManager{}
}

func (s *SymlinkManager) SwitchVersion(category, targetVersionDir string) error {
	baseDir := config.GetBaseDir()
	binDir := filepath.Join(baseDir, "bin", category)
	currentLink := filepath.Join(binDir, "current")

	// targetVersionDir is relative to binDir, e.g., "php-8.2.0"
	targetPath := filepath.Join(binDir, targetVersionDir)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("target version directory does not exist: %s", targetPath)
	}

	// Remove existing symlink or folder
	if _, err := os.Lstat(currentLink); err == nil {
		// On windows, Junctions are removed with os.Remove if they are empty
		// but sometimes os.RemoveAll is safer or use cmd /c rmdir
		os.Remove(currentLink)
	}

	if runtime.GOOS == "windows" {
		// Use Directory Junction on Windows (mklink /J)
		// It doesn't require admin privileges and is very portable.
		cmd := utils.Executor.Command("cmd", "/c", "mklink", "/J", currentLink, targetPath)
		return cmd.Run()
	}

	return os.Symlink(targetPath, currentLink)
}

