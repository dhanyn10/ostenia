package service

import (
	"ostenia/internal/backend/interfaces"
	"path/filepath"
	"strings"
)

// GetPath retrieves the current PATH environment variable from specific target (User or Machine).
func GetPath(sys interfaces.System, target string) (string, error) {
	return sys.GetPath(target)
}

// SetPath sets the PATH environment variable. If target is Machine, it triggers a native Windows UAC prompt.
func SetPath(sys interfaces.System, path string, target string) error {
	return sys.SetPath(path, target)
}

// UpdatePHPPath manages PHP entries in the USER PATH.
func UpdatePHPPath(sys interfaces.System, phpPath string, add bool) error {
	currentPath, err := GetPath(sys, "User")
	if err != nil {
		return err
	}
	paths := strings.Split(currentPath, ";")
	var newPaths []string
	normalizedTarget := filepath.Clean(strings.ToLower(phpPath))
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		cleanP := filepath.Clean(strings.ToLower(trimmed))
		// Identify and remove existing Ostenia PHP paths
		if cleanP == normalizedTarget || (strings.Contains(cleanP, "ostenia") && strings.Contains(cleanP, "php")) {
			continue
		}
		newPaths = append(newPaths, trimmed)
	}
	if add {
		newPaths = append([]string{phpPath}, newPaths...)
	}
	return SetPath(sys, strings.Join(newPaths, ";"), "User")
}

// UpdateNodePath manages Node.js entries in the SYSTEM (Machine) PATH.
func UpdateNodePath(sys interfaces.System, nodePath string, add bool) error {
	return updateSystemComponentPath(sys, nodePath, "node", add)
}

// UpdatePythonPath manages Python entries in the SYSTEM (Machine) PATH.
func UpdatePythonPath(sys interfaces.System, pythonPath string, add bool) error {
	err := updateSystemComponentPath(sys, pythonPath, "python", add)
	if err != nil {
		return err
	}
	scriptsPath := filepath.Join(pythonPath, "Scripts")
	return updateSystemComponentPath(sys, scriptsPath, "python-scripts", add)
}

// updateSystemComponentPath handles generic system-level PATH management for components.
func updateSystemComponentPath(sys interfaces.System, targetPath string, keyword string, add bool) error {
	currentPath, err := GetPath(sys, "Machine")
	if err != nil {
		return err
	}
	paths := strings.Split(currentPath, ";")
	var newPaths []string
	normalizedTarget := filepath.Clean(strings.ToLower(targetPath))
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		cleanP := filepath.Clean(strings.ToLower(trimmed))
		// Identify Ostenia component path and filter it out
		if cleanP == normalizedTarget || (strings.Contains(cleanP, "ostenia") && strings.Contains(cleanP, keyword)) {
			continue
		}
		newPaths = append(newPaths, trimmed)
	}
	if add {
		newPaths = append([]string{targetPath}, newPaths...)
	}
	return SetPath(sys, strings.Join(newPaths, ";"), "Machine")
}

// IsPathInUserPath checks if a path exists in the User PATH.
func IsPathInUserPath(sys interfaces.System, targetPath string) bool {
	current, _ := GetPath(sys, "User")
	return pathExistsInString(current, targetPath)
}

// IsPathInSystemPath checks if a path exists in the Machine (System) PATH.
func IsPathInSystemPath(sys interfaces.System, targetPath string) bool {
	current, _ := GetPath(sys, "Machine")
	return pathExistsInString(current, targetPath)
}

// pathExistsInString verifies if a specific targetPath is present in a semicolon-separated string.
func pathExistsInString(pathString, targetPath string) bool {
	if pathString == "" {
		return false
	}
	normalizedTarget := filepath.Clean(strings.ToLower(targetPath))
	paths := strings.Split(pathString, ";")
	for _, p := range paths {
		if filepath.Clean(strings.ToLower(strings.TrimSpace(p))) == normalizedTarget {
			return true
		}
	}
	return false
}

// NotifyEnvironmentUpdate broadcasts WM_SETTINGCHANGE to all windows to refresh environment variables.
func NotifyEnvironmentUpdate() {
	notifyEnvironmentUpdate()
}
