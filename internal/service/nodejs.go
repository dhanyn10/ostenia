package service

import (
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
)

func GetNodeVersion(currentPath string) (string, error) {
	nodeExe := filepath.Join(currentPath, "node.exe")
	if _, err := os.Stat(nodeExe); err != nil {
		return "", err
	}
	cmd := exec.Command(nodeExe, "-v")
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
