package download

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// EnsureCurrentLink manages the 'current' symlink/junction for a component.
func EnsureCurrentLink(baseDir string, targetPath string, componentName string) error {
	dir := filepath.Dir(targetPath)
	category := filepath.Base(dir) // e.g., "php"

	currentLink := filepath.Join(filepath.Dir(dir), category, "current")
	targetAbs := targetPath

	// If no version subfolder (e.g. "bin/heidisql/heidisql.exe"), no symlink
	if category == "bin" {
		return nil
	}

	fmt.Printf("[Symlinker] Linking %s -> %s\n", currentLink, targetAbs)

	// Remove existing link if it exists
	if _, err := os.Lstat(currentLink); err == nil {
		os.Remove(currentLink)
	}

	// Create symlink/junction
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "mklink", "/J", currentLink, targetAbs)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("could not create junction for %s: %v", componentName, err)
		}
	} else {
		err := os.Symlink(targetAbs, currentLink)
		if err != nil {
			return fmt.Errorf("could not create symlink for %s: %v", componentName, err)
		}
	}

	return nil
}
