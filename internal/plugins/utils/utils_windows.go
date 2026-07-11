//go:build windows

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

// DetectHeidiSQLInstallation checks if HeidiSQL is installed in the system and returns its path and uninstaller.
func DetectHeidiSQLInstallation() (string, string) {
	if exe, uninst := checkHeidiSQLRegistry(); exe != "" {
		return exe, uninst
	}
	if exe, uninst := checkHeidiSQLCommonPaths(); exe != "" {
		return exe, uninst
	}
	return checkHeidiSQLPath()
}

func checkHeidiSQLRegistry() (string, string) {
	keys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\HeidiSQL_is1`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\HeidiSQL_is1`,
	}
	for _, k := range keys {
		regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, k, registry.QUERY_VALUE)
		if err != nil {
			regKey, _ = registry.OpenKey(registry.CURRENT_USER, k, registry.QUERY_VALUE)
		}
		if regKey != 0 {
			exe, _, _ := regKey.GetStringValue("DisplayIcon")
			uninst, _, _ := regKey.GetStringValue("UninstallString")
			regKey.Close()
			if exe != "" {
				return strings.Trim(exe, "\""), strings.Trim(uninst, "\"")
			}
		}
	}
	return "", ""
}

func checkHeidiSQLCommonPaths() (string, string) {
	commonPaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "HeidiSQL", "heidisql.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "HeidiSQL", "heidisql.exe"),
	}
	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, getHeidiSQLUninstaller(p)
		}
	}
	return "", ""
}

func checkHeidiSQLPath() (string, string) {
	cmd := exec.Command("cmd", "/c", "where heidisql.exe") // NOSONAR
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err == nil {
		p := strings.TrimSpace(strings.Split(string(out), "\r\n")[0])
		if p != "" {
			return p, getHeidiSQLUninstaller(p)
		}
	}
	return "", ""
}

func getHeidiSQLUninstaller(exePath string) string {
	dir := filepath.Dir(exePath)
	uninst := filepath.Join(dir, "unins000.exe")
	if _, err := os.Stat(uninst); err == nil {
		return uninst
	}
	uninst = filepath.Join(dir, "uninstall.exe")
	if _, err := os.Stat(uninst); err == nil {
		return uninst
	}
	return ""
}
