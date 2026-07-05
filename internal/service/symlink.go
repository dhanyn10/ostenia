package service

import (
	"fmt"
	"os"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"path/filepath"
)

type SymlinkManager struct {
	system interfaces.System
}

func NewSymlinkManager(system interfaces.System) *SymlinkManager {
	if system == nil {
		system = NewWindowsSystem(nil)
	}
	return &SymlinkManager{system: system}
}

func (s *SymlinkManager) SwitchVersion(category string, targetVersionDir string) error {
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

	return s.system.CreateSymlink(targetPath, currentLink)
}

