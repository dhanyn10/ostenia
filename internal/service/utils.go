package service

import (
	"fmt"
	"os"
	"os/exec"
	"ostenia/internal/network"
	"runtime"
	"strings"
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

func AddHostWithElevation(ip string, hostname string) error {
	if IsAdmin() {
		return network.AddHost(ip, hostname)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// We'll use a small trick: run the app itself with a special flag/command
	// or ideally a small helper. Since we're in a Wails app, we can use the same exe
	// if we handle the command line args in main.go

	if runtime.GOOS == "windows" {
		// Use powershell to start a process as admin to run a command that adds the host
		// We'll call ourselves with a special flag --add-host
		args := fmt.Sprintf("--add-host %s %s", ip, hostname)
		return exec.Command("cmd", "/c", "powershell", "Start-Process", "-FilePath", fmt.Sprintf("'%s'", exe), "-ArgumentList", fmt.Sprintf("'%s'", args), "-Verb", "runas", "-WindowStyle", "Hidden").Run()
	}

	return fmt.Errorf("elevation required but not supported on this platform")
}
