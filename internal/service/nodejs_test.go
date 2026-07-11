package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"testing"
)

func TestGetNodeVersion(t *testing.T) {
	tempDir := t.TempDir()
	nodeExe := filepath.Join(tempDir, "node.exe")
	_ = os.WriteFile(nodeExe, []byte(""), 0755)

	oldExecutor := utils.Executor
	defer func() { utils.Executor = oldExecutor }()

	utils.Executor = &testutil.MockExecutor{
		Output: "v18.16.0",
	}

	version, err := GetNodeVersion(tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != "v18.16.0" {
		t.Errorf("Expected v18.16.0, got %s", version)
	}

	// Test error case
	_ = os.Remove(nodeExe)
	_, err = GetNodeVersion(tempDir)
	if err == nil {
		t.Error("Expected error for missing node.exe, got nil")
	}
}
