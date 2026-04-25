package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Terminal struct {
	WorkingDir string
	Env        []string
}

func NewTerminal(workingDir string, env []string) *Terminal {
	return &Terminal{
		WorkingDir: workingDir,
		Env:        env,
	}
}

func (t *Terminal) Open(terminalType string) error {
	var cmd *exec.Cmd

	if runtime.GOOS != "windows" {
		cmd = exec.Command("bash")
		cmd.Dir = t.WorkingDir
		cmd.Env = t.Env
		return cmd.Start()
	}

	switch terminalType {
	case "powershell":
		cmd = exec.Command("cmd.exe", "/C", "start", "powershell.exe", "-NoExit", "-Command", "Set-Location '"+t.WorkingDir+"'; $Host.UI.RawUI.WindowTitle = 'Ostenia PowerShell'")
	case "gitbash":
		bashPaths := []string{
			`C:\Program Files\Git\bin\bash.exe`,
			`C:\Program Files\x86)\Git\bin\bash.exe`,
			os.Getenv("USERPROFILE") + `\AppData\Local\Programs\Git\bin\bash.exe`,
		}

		var bashPath string
		for _, p := range bashPaths {
			if _, err := os.Stat(p); err == nil {
				bashPath = p
				break
			}
		}

		if bashPath != "" {
			cmd = exec.Command("cmd.exe", "/C", "start", "", bashPath, "--login", "-i")
		} else {
			cmd = exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", "title Ostenia Terminal (Git Bash not found)")
		}
	default: // cmd
		cmd = exec.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", "title Ostenia Terminal")
	}

	cmd.Dir = t.WorkingDir
	cmd.Env = t.Env
	return cmd.Start()
}

func (t *Terminal) Start() error {
	return t.Open("cmd")
}

// OpenExplorer opens the given path in the system's file explorer.
func OpenExplorer(path string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Using 'start' instead of 'explorer' directly to avoid the "exit status 1" quirk of explorer.exe
		cmd = exec.Command("cmd", "/c", "start", "", filepath.FromSlash(path))
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", path)
	} else {
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start() // Use Start() instead of Run() to avoid waiting for exit status
}
