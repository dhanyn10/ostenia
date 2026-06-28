package utils

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetSystemDirectory(t *testing.T) {
	dir := GetSystemDirectory()
	if dir == "" {
		t.Error("Expected non-empty system directory")
	}
	if runtime.GOOS == "windows" {
		if !strings.Contains(strings.ToLower(dir), "system32") {
			t.Errorf("Expected windows system directory to contain system32, got %s", dir)
		}
	} else {
		if dir != "/usr/bin" {
			t.Errorf("Expected unix system directory to be /usr/bin, got %s", dir)
		}
	}
}

func TestSafeEnv(t *testing.T) {
	env := SafeEnv()
	foundPath := false
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			foundPath = true
			break
		}
	}
	if !foundPath {
		t.Error("Expected PATH to be present in SafeEnv")
	}
}
