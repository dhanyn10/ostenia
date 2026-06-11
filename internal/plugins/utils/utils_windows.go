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
	// 1. Try Registry first (Standard Windows way)
	keys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\HeidiSQL_is1`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\HeidiSQL_is1`,
	}

	for _, k := range keys {
		regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, k, registry.QUERY_VALUE)
		if err != nil {
			regKey, err = registry.OpenKey(registry.CURRENT_USER, k, registry.QUERY_VALUE)
		}
		if err == nil {
			displayIcon, _, err := regKey.GetStringValue("DisplayIcon")
			if err == nil {
				exePath = strings.Trim(displayIcon, "\"")
			}
			uninstString, _, err := regKey.GetStringValue("UninstallString")
			if err == nil {
				uninstaller = strings.Trim(uninstString, "\"")
			}
			regKey.Close()
			if exePath != "" {
				return
			}
		}
	}

	// 2. Check common installation paths
	commonPaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "HeidiSQL", "heidisql.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "HeidiSQL", "heidisql.exe"),
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			exePath = p
			// Inno Setup (HeidiSQL) uses unins000.exe
			uninstaller = filepath.Join(filepath.Dir(p), "unins000.exe")
			if _, err := os.Stat(uninstaller); err != nil {
				uninstaller = filepath.Join(filepath.Dir(p), "uninstall.exe")
				if _, err := os.Stat(uninstaller); err != nil {
					uninstaller = ""
				}
			}
			return
		}
	}

	// 2. Fallback to Registry check via WMIC or cmd
	// We use "where heidisql" as a simple fallback if it's in the PATH
	cmd := exec.Command("cmd", "/c", "where heidisql.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err == nil {
		p := strings.TrimSpace(strings.Split(string(out), "\r\n")[0])
		if p != "" {
			exePath = p
			uninstaller = filepath.Join(filepath.Dir(p), "unins000.exe")
			if _, err := os.Stat(uninstaller); err != nil {
				uninstaller = filepath.Join(filepath.Dir(p), "uninstall.exe")
				if _, err := os.Stat(uninstaller); err != nil {
					uninstaller = ""
				}
			}
			return
		}
	}

	return "", ""
}
