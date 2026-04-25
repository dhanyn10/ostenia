package service

import (
	"os"
	"os/exec"
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
		// Open PowerShell in a new window
		cmd = exec.Command("cmd.exe", "/C", "start", "powershell.exe", "-NoExit", "-Command", "Set-Location '"+t.WorkingDir+"'; $Host.UI.RawUI.WindowTitle = 'Ostenia PowerShell'")
	case "gitbash":
		// Try to find Git Bash in common locations
		bashPaths := []string{
			`C:\Program Files\Git\bin\bash.exe`,
			`C:\Program Files (x86)\Git\bin\bash.exe`,
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
			// Fallback to CMD if Git Bash not found
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
