//go:build windows

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
)

func TestSetHideWindow(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "echo", "test")
	SetHideWindow(cmd)
	if cmd.SysProcAttr == nil || !cmd.SysProcAttr.HideWindow {
		t.Error("Expected HideWindow to be true")
	}

	cmd2 := exec.Command("cmd", "/c", "echo", "test2")
	cmd2.SysProcAttr = &syscall.SysProcAttr{}
	SetHideWindow(cmd2)
	if cmd2.SysProcAttr == nil || !cmd2.SysProcAttr.HideWindow {
		t.Error("Expected HideWindow to be true when SysProcAttr was pre-initialized")
	}
}

func TestGetHeidiSQLUninstaller(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. With unins000.exe
	dir1 := filepath.Join(tmpDir, "app1")
	if err := os.MkdirAll(dir1, 0700); err != nil {
		t.Fatal(err)
	}
	exePath1 := filepath.Join(dir1, "heidisql.exe")
	uninst1 := filepath.Join(dir1, "unins000.exe")
	_ = os.WriteFile(exePath1, []byte("exe"), 0600)
	_ = os.WriteFile(uninst1, []byte("uninst"), 0600)

	got := getHeidiSQLUninstaller(exePath1)
	if got != uninst1 {
		t.Errorf("Expected %q, got %q", uninst1, got)
	}

	// 2. With uninstall.exe
	dir2 := filepath.Join(tmpDir, "app2")
	if err := os.MkdirAll(dir2, 0700); err != nil {
		t.Fatal(err)
	}
	exePath2 := filepath.Join(dir2, "heidisql.exe")
	uninst2 := filepath.Join(dir2, "uninstall.exe")
	_ = os.WriteFile(exePath2, []byte("exe"), 0600)
	_ = os.WriteFile(uninst2, []byte("uninst"), 0600)

	got = getHeidiSQLUninstaller(exePath2)
	if got != uninst2 {
		t.Errorf("Expected %q, got %q", uninst2, got)
	}

	// 3. Neither exists
	dir3 := filepath.Join(tmpDir, "app3")
	if err := os.MkdirAll(dir3, 0700); err != nil {
		t.Fatal(err)
	}
	exePath3 := filepath.Join(dir3, "heidisql.exe")
	_ = os.WriteFile(exePath3, []byte("exe"), 0600)

	got = getHeidiSQLUninstaller(exePath3)
	if got != "" {
		t.Errorf("Expected empty string, got %q", got)
	}
}

func TestCheckHeidiSQLCommonPaths(t *testing.T) {
	tmpDir := t.TempDir()
	origProgFiles := os.Getenv("ProgramFiles")
	defer os.Setenv("ProgramFiles", origProgFiles)

	heidiDir := filepath.Join(tmpDir, "HeidiSQL")
	if err := os.MkdirAll(heidiDir, 0700); err != nil {
		t.Fatal(err)
	}
	exePath := filepath.Join(heidiDir, "heidisql.exe")
	_ = os.WriteFile(exePath, []byte("exe"), 0600)

	os.Setenv("ProgramFiles", tmpDir)

	exe, _ := checkHeidiSQLCommonPaths()
	if exe != exePath {
		t.Errorf("Expected %q, got %q", exePath, exe)
	}

	// Test when path does not exist
	os.Setenv("ProgramFiles", filepath.Join(tmpDir, "empty"))
	exeNotFound, _ := checkHeidiSQLCommonPaths()
	if exeNotFound != "" {
		t.Errorf("Expected empty string when not found, got %q", exeNotFound)
	}
}

func TestDetectHeidiSQLInstallation(t *testing.T) {
	exe, uninst := DetectHeidiSQLInstallation()
	t.Logf("Detected HeidiSQL: exe=%q, uninst=%q", exe, uninst)
}
