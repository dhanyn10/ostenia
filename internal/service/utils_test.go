package service

import (
	"runtime"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	// We can't easily force admin rights in a test, but we can check if it runs without crashing
	// and returns a boolean.
	res := IsAdmin()
	if runtime.GOOS == "linux" {
		// In GitHub Actions it might be root
		t.Logf("IsAdmin on Linux: %v", res)
	} else if runtime.GOOS == "windows" {
		t.Logf("IsAdmin on Windows: %v", res)
	}
}
