package plugins

import (
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
)

func CreateSymlink(oldname, newname string) error {
	os.Remove(newname)
	if runtime.GOOS == "windows" {
		// Use Junction on Windows as it doesn't require admin privileges
		cmd := utils.Executor.Command("cmd", "/c", "mklink", "/J", newname, oldname)
		utils.SetHideWindow(cmd)
		return cmd.Run()
	}
	return os.Symlink(oldname, newname)
}

func ResolveSymlink(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
