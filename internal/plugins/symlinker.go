package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func CreateSymlink(oldname, newname string) error {
	os.Remove(newname)
	if runtime.GOOS == "windows" {
		// Use Junction on Windows as it doesn't require admin privileges
		cmd := exec.Command("cmd", "/c", "mklink", "/J", newname, oldname)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		return cmd.Run()
	}
	return os.Symlink(oldname, newname)
}

func ResolveSymlink(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
