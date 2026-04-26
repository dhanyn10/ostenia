package service

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

// UpdateUserPath adds or removes Ostenia-related PHP directories from the Windows User PATH
func UpdateUserPath(targetPath string, add bool) error {
	// Get current User PATH using PowerShell
	getCmd := exec.Command("powershell", "-Command", "[Environment]::GetEnvironmentVariable('Path', 'User')")
	out, err := getCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get user path: %w", err)
	}

	currentPath := strings.TrimSpace(string(out))
	paths := strings.Split(currentPath, ";")

	var newPaths []string
	found := false

	// Normalize target path for comparison
	normalizedTarget := filepath.Clean(strings.ToLower(targetPath))

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" { continue }

		cleanP := filepath.Clean(strings.ToLower(p))

		// If we are removing (toggle off)
		if !add {
			// Check if this path is exactly the target OR is any path containing both 'ostenia' and 'php'
			// This ensures we clean up Ostenia PHP related paths thoroughly
			isOsteniaPHP := strings.Contains(cleanP, "ostenia") && strings.Contains(cleanP, "php")
			if cleanP == normalizedTarget || isOsteniaPHP {
				found = true
				continue // Skip/Remove this path
			}
		} else {
			// If we are adding (toggle on), check if it already exists to avoid duplicates
			if cleanP == normalizedTarget {
				found = true
			}
		}
		newPaths = append(newPaths, p)
	}

	if add && !found {
		newPaths = append(newPaths, targetPath)
	}

	// Only update if there was an actual change
	if (add && !found) || (!add && found) {
		finalPath := strings.Join(newPaths, ";")
		setCmd := exec.Command("powershell", "-Command", fmt.Sprintf("[Environment]::SetEnvironmentVariable('Path', '%s', 'User')", finalPath))
		err = setCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to set user path: %w", err)
		}
		NotifyEnvironmentUpdate()
	}

	return nil
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
