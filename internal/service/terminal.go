package service

import (
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
)

var RuntimeGOOS = runtime.GOOS

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

	if RuntimeGOOS != "windows" {
		cmd = utils.Executor.Command("bash")
		cmd.Dir = t.WorkingDir
		cmd.Env = t.Env
		return cmd.Start()
	}

	switch terminalType {
	case "powershell":
		cmd = utils.Executor.Command("cmd.exe", "/C", "start", "powershell.exe", "-NoExit", "-Command", "Set-Location '"+t.WorkingDir+"'; $Host.UI.RawUI.WindowTitle = 'Ostenia PowerShell'")
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
			cmd = utils.Executor.Command("cmd.exe", "/C", "start", "", bashPath, "--login", "-i")
		} else {
			cmd = utils.Executor.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", "title Ostenia Terminal (Git Bash not found)")
		}
	default: // cmd
		cmd = utils.Executor.Command("cmd.exe", "/C", "start", "cmd.exe", "/K", "title Ostenia Terminal")
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
	if RuntimeGOOS == "windows" {
		// Using 'start' instead of 'explorer' directly to avoid the "exit status 1" quirk of explorer.exe
		cmd = utils.Executor.Command("cmd", "/c", "start", "", filepath.FromSlash(path))
	} else if RuntimeGOOS == "darwin" {
		cmd = utils.Executor.Command("open", path)
	} else {
		cmd = utils.Executor.Command("xdg-open", path)
	}
	return cmd.Start() // Use Start() instead of Run() to avoid waiting for exit status
}
