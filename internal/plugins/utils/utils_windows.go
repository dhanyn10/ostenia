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
func DetectHeidiSQLInstallation() (exePath string, uninstaller string) {
	if exe, uninst := detectByRegistry(); exe != "" {
		return exe, uninst
	}
	if exe, uninst := detectByCommonPaths(); exe != "" {
		return exe, uninst
	}
	return detectByPath()
}

func detectByRegistry() (string, string) {
	keys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\HeidiSQL_is1`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\HeidiSQL_is1`,
	}

	for _, k := range keys {
		if exe, uninst := checkRegistryKey(registry.LOCAL_MACHINE, k); exe != "" {
			return exe, uninst
		}
		if exe, uninst := checkRegistryKey(registry.CURRENT_USER, k); exe != "" {
			return exe, uninst
		}
	}
	return "", ""
}

func checkRegistryKey(root registry.Key, path string) (string, string) {
	regKey, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return "", ""
	}
	defer regKey.Close()

	displayIcon, _, err := regKey.GetStringValue("DisplayIcon")
	exePath := strings.Trim(displayIcon, "\"")

	uninstString, _, err := regKey.GetStringValue("UninstallString")
	uninstaller := strings.Trim(uninstString, "\"")

	return exePath, uninstaller
}

func detectByCommonPaths() (string, string) {
	commonPaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "HeidiSQL", "heidisql.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "HeidiSQL", "heidisql.exe"),
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p, getUninstallerPath(p)
		}
	}
	return "", ""
}

func detectByPath() (string, string) {
	cmd := exec.Command("cmd", "/c", "where heidisql.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return "", ""
	}

	p := strings.TrimSpace(strings.Split(string(out), "\r\n")[0])
	if p != "" {
		return p, getUninstallerPath(p)
	}
	return "", ""
}

func getUninstallerPath(exePath string) string {
	dir := filepath.Dir(exePath)
	uninst000 := filepath.Join(dir, "unins000.exe")
	if _, err := os.Stat(uninst000); err == nil {
		return uninst000
	}
	uninst := filepath.Join(dir, "uninstall.exe")
	if _, err := os.Stat(uninst); err == nil {
		return uninst
	}
	return ""
}
