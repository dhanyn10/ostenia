package php

import (
	"context"
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"testing"
)

func TestGetModuleVersion_Composer(t *testing.T) {
	tempDir := t.TempDir()
	phpExe := filepath.Join(tempDir, "php.exe")
	composerPhar := filepath.Join(tempDir, "composer.phar")

	os.WriteFile(phpExe, []byte(""), 0755)
	os.WriteFile(composerPhar, []byte(""), 0644)

	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	mock := &mockExecutor{
		responses: map[string]string{
			"composer.phar": "Composer version 2.6.5 2023-10-06 10:11:52",
		},
	}
	utils.Executor = mock

	v := GetModuleVersion("Composer", tempDir)
	if v != "2.6.5" {
		t.Errorf("Expected 2.6.5, got %s", v)
	}
}

func TestInstallModule_Composer(t *testing.T) {
	tempDir := t.TempDir()

	origClient := utils.Client
	defer func() { utils.Client = origClient }()
	utils.Client = &mockHTTPClient{content: "mock composer phar"}

	err := InstallModule(context.Background(), nil, "Composer", tempDir, func(name string, pct float64, status string) {})
	if err != nil {
		t.Fatalf("InstallModule failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "composer.phar")); os.IsNotExist(err) {
		t.Error("composer.phar not created")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "composer.bat")); os.IsNotExist(err) {
		t.Error("composer.bat not created")
	}
}

func TestInstallModule_Unknown(t *testing.T) {
	err := InstallModule(context.Background(), nil, "Unknown", "path", nil)
	if err == nil {
		t.Error("Expected error for unknown module")
	}
}
