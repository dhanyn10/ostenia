package openssl

import (
	_ "embed"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed openssl.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	version := "4.0.0"
	arch := utils.GetSystemArch()
	var url string
	if arch == "x64" {
		url = "https://slproweb.com/download/Win64OpenSSL_Light-4_0_0.exe"
	} else {
		url = "https://slproweb.com/download/Win32OpenSSL_Light-4_0_0.exe"
	}
	return []string{version}, map[string]string{version: url}
}

func GetIcon() string {
	return iconSVG
}

// DetectInstalledVersion verifies an OpenSSL executable by running `<path> version`.
func DetectInstalledVersion() string {
	if runtime.GOOS == "windows" {
		for _, path := range findExecutables() {
			if version := versionFromExecutable(path); version != "" {
				return version
			}
		}
	}

	return versionFromExecutable("openssl")
}

func findExecutables() []string {
	paths := []string{}
	seen := map[string]bool{}

	addPath := func(path string) {
		path = strings.Trim(path, " \t\r\n\"")
		if path == "" {
			return
		}
		key := strings.ToLower(path)
		if seen[key] {
			return
		}
		if _, err := os.Stat(path); err == nil {
			seen[key] = true
			paths = append(paths, path)
		}
	}

	for _, cmd := range []*exec.Cmd{
		utils.Executor.Command("cmd", "/d", "/c", "where openssl"),
		utils.Executor.Command("where.exe", "openssl"),
	} {
		utils.SetHideWindow(cmd)
		if out, err := cmd.Output(); err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				addPath(line)
			}
		}
	}

	filepath.Walk(filepath.Join(config.GetBaseDir(), "bin"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if strings.EqualFold(info.Name(), "openssl.exe") {
			addPath(path)
		}
		return nil
	})

	return paths
}

func versionFromExecutable(exePath string) string {
	cmd := utils.Executor.Command(exePath, "version")
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	parts := strings.Split(string(out), " ")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
