//go:build !windows

package service

func notifyEnvironmentUpdate() {
	// No-op: broadcasting environment updates via WM_SETTINGCHANGE is specific to Windows.
}

func findOsteniaPIDsWindows(exeName, binPath string) []int {
	return []int{}
}
