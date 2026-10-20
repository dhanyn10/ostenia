//go:build windows

package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindOsteniaPIDsWindows(t *testing.T) {
	// 1. Empty binPath
	pids := findOsteniaPIDsWindows("anything.exe", "")
	if len(pids) != 0 {
		t.Errorf("Expected 0 pids for empty binPath, got %d", len(pids))
	}

	// 2. Non-matching binPath
	pids = findOsteniaPIDsWindows("anything.exe", "Z:\\non-existent-directory-12345")
	if len(pids) != 0 {
		t.Errorf("Expected 0 pids for non-existent path, got %d", len(pids))
	}

	// 3. Find current running test process
	exePath, err := os.Executable()
	if err != nil {
		t.Fatalf("Failed to get executable: %v", err)
	}

	exeName := filepath.Base(exePath)
	binDir := filepath.Dir(exePath)

	pids = findOsteniaPIDsWindows(exeName, binDir)
	currentPid := os.Getpid()
	found := false
	for _, pid := range pids {
		if pid == currentPid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected current PID %d in results for %s in %s, got: %v", currentPid, exeName, binDir, pids)
	}
}
