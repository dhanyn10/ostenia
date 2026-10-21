//go:build !windows

package service

func notifyEnvironmentUpdate() {}

func findOsteniaPIDsWindows(exeName, binPath string) []int {
	return []int{}
}
