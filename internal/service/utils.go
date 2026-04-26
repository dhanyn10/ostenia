package service

import (
	"os"
	"runtime"
)

// IsAdmin checks if the current process has administrative privileges
func IsAdmin() bool {
	if runtime.GOOS == "windows" {
		// On Windows, checking if we can open the physical drive is a common way to check for Admin rights
		f, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		if err != nil {
			return false
		}
		f.Close()
		return true
	}
	// For other OS, check UID
	return os.Geteuid() == 0
}
