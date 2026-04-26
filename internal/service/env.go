package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// GetPath retrieves the current PATH environment variable from specific target (User or Machine)
func GetPath(target string) (string, error) {
	getCmd := exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("[Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::%s)", target))
	getCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := getCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get %s path: %w", target, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SetPath sets the PATH environment variable. If target is Machine, it triggers a native Windows UAC prompt.
func SetPath(path string, target string) error {
	if target == "Machine" && !IsAdmin() {
		// Use a temporary powershell script to avoid quoting hell and ensure UAC triggers
		scriptContent := fmt.Sprintf("[Environment]::SetEnvironmentVariable('Path', '%s', [EnvironmentVariableTarget]::Machine)", path)
		tmpScript := filepath.Join(os.TempDir(), "ostenia_set_path.ps1")
		os.WriteFile(tmpScript, []byte(scriptContent), 0644)
		defer os.Remove(tmpScript)

		// Trigger UAC using Start-Process with RunAs verb
		// We DON'T hide the window here to ensure the prompt is visible
		args := fmt.Sprintf("-NoProfile -ExecutionPolicy Bypass -File \"%s\"", tmpScript)
		elevatedCmd := fmt.Sprintf("Start-Process powershell -ArgumentList '%s' -Verb RunAs -Wait", args)

		cmd := exec.Command("powershell", "-NoProfile", "-Command", elevatedCmd)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("UAC prompt denied or failed: %w", err)
		}
	} else {
		// Set directly if User target or already Admin
		script := fmt.Sprintf("[Environment]::SetEnvironmentVariable('Path', '%s', [EnvironmentVariableTarget]::%s)", path, target)
		cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to set %s path: %w", target, err)
		}
	}

	NotifyEnvironmentUpdate()
	return nil
}

// UpdatePHPPath manages PHP entries in the USER PATH
func UpdatePHPPath(phpPath string, add bool) error {
	currentPath, err := GetPath("User")
	if err != nil { return err }

	paths := strings.Split(currentPath, ";")
	var newPaths []string
	normalizedTarget := filepath.Clean(strings.ToLower(phpPath))

	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" { continue }
		cleanP := filepath.Clean(strings.ToLower(trimmed))
		if cleanP == normalizedTarget || (strings.Contains(cleanP, "ostenia") && strings.Contains(cleanP, "php")) {
			continue
		}
		newPaths = append(newPaths, trimmed)
	}

	if add {
		newPaths = append([]string{phpPath}, newPaths...)
	}

	return SetPath(strings.Join(newPaths, ";"), "User")
}

// UpdateNodePath manages Node.js entries in the SYSTEM (Machine) PATH
func UpdateNodePath(nodePath string, add bool) error {
	currentPath, err := GetPath("Machine")
	if err != nil { return err }

	paths := strings.Split(currentPath, ";")
	var newPaths []string
	normalizedTarget := filepath.Clean(strings.ToLower(nodePath))

	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" { continue }
		cleanP := filepath.Clean(strings.ToLower(trimmed))
		if cleanP == normalizedTarget || (strings.Contains(cleanP, "ostenia") && (strings.Contains(cleanP, "node") || strings.Contains(cleanP, "npm"))) {
			continue
		}
		newPaths = append(newPaths, trimmed)
	}

	if add {
		newPaths = append([]string{nodePath}, newPaths...)
	}

	return SetPath(strings.Join(newPaths, ";"), "Machine")
}

// IsPathInUserPath checks User PATH
func IsPathInUserPath(targetPath string) bool {
	current, _ := GetPath("User")
	return pathExistsInString(current, targetPath)
}

// IsPathInSystemPath checks Machine (System) PATH
func IsPathInSystemPath(targetPath string) bool {
	current, _ := GetPath("Machine")
	return pathExistsInString(current, targetPath)
}

func pathExistsInString(pathString, targetPath string) bool {
	if pathString == "" { return false }
	normalizedTarget := filepath.Clean(strings.ToLower(targetPath))
	paths := strings.Split(pathString, ";")
	for _, p := range paths {
		if filepath.Clean(strings.ToLower(strings.TrimSpace(p))) == normalizedTarget {
			return true
		}
	}
	return false
}

// NotifyEnvironmentUpdate broadcasts WM_SETTINGCHANGE to all windows
func NotifyEnvironmentUpdate() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessage := user32.NewProc("SendMessageTimeoutW")

	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)

	envStr, _ := syscall.UTF16PtrFromString("Environment")
	sendMessage.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(envStr)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000),
		0,
	)
}
