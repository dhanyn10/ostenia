package service

import (
	"fmt"
	"os"
	"ostenia/internal/config"
	"path/filepath"
)

type SymlinkManager struct{}

func NewSymlinkManager() *SymlinkManager {
	return &SymlinkManager{}
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
		err = os.Remove(currentLink)
		if err != nil {
			return err
		}
	}

	// Create new symlink
	// On Windows, os.Symlink requires SeCreateSymbolicLinkPrivilege or Admin
	// Since we are running as Admin, it should work.
	return os.Symlink(targetPath, currentLink)
}
