package service

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
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

func RunMeAsAdmin() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	verb := "runas"
	args := strings.Join(os.Args[1:], " ")

	if runtime.GOOS == "windows" {
		cmd := fmt.Sprintf("/c start \"\" \"%s\" %s", exe, args)
		return exec.Command("cmd", "/c", "powershell", "Start-Process", "-FilePath", fmt.Sprintf("'%s'", exe), "-ArgumentList", fmt.Sprintf("'%s'", args), "-Verb", verb).Run()
	}

	return fmt.Errorf("elevation not supported on this platform")
}

func ElevateAndExit() {
	err := RunMeAsAdmin()
	if err == nil {
		os.Exit(0)
	}
}
