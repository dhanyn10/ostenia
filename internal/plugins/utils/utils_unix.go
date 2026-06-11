//go:build !windows

package utils

// DetectHeidiSQLInstallation is a fallback for non-Windows platforms.
func DetectHeidiSQLInstallation() (exePath string, uninstaller string) {
	return "", ""
}
