package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
)

func GetPythonVersion(currentPath string) (string, error) {
	exePath := filepath.Join(currentPath, "python.exe")
	if _, err := os.Stat(exePath); err != nil {
		return "", err
	}
	cmd := utils.Executor.Command(exePath, "--version")
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
