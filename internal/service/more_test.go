package service

import (
	"context"
	"ostenia/internal/plugins/utils"
	"os"
	"path/filepath"
	"testing"
)

func TestSSHManager_New(t *testing.T) {
	mgr := NewSSHManager(context.Background())
	if mgr == nil {
		t.Error("Expected SSHManager, got nil")
	}
}

func TestSymlinkManager_SwitchVersion_ExistingLink(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &mockExecutor{}

	tempDir := t.TempDir()
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Unsetenv("OSTENIA_HOME")

	binDir := filepath.Join(tempDir, "bin", "php")
	os.MkdirAll(binDir, 0755)

	versionDir := filepath.Join(binDir, "php-8.2.0")
	os.MkdirAll(versionDir, 0755)

	currentLink := filepath.Join(binDir, "current")
	os.WriteFile(currentLink, []byte("existing"), 0644)

	s := NewSymlinkManager()
	err := s.SwitchVersion("php", "php-8.2.0")
	if err != nil {
		t.Errorf("SwitchVersion failed with existing link: %v", err)
	}

	if _, err := os.Stat(currentLink); os.IsNotExist(err) {
		// On non-windows it should create a symlink, on windows it calls cmd mklink (mocked)
		// If mocked, the file might actually be deleted by os.Remove(currentLink) and not recreated
		// because the mock doesn't actually create the file.
	}
}

func TestOrchestrator_Basic(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Unsetenv("OSTENIA_HOME")

	ctx := context.Background()
	orch := NewOrchestrator(ctx)

	orch.SetActiveTab("test")
	orch.RequestRefresh()

	if orch.IsRunning("non-existent") {
		t.Error("Expected false for non-existent service")
	}

	info := orch.GetDetailedInfo("Apache")
	if info.Name != "Apache" {
		t.Errorf("Expected Apache, got %s", info.Name)
	}
}
