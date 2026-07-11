package service

import (
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"testing"
)

func TestGetPythonVersion(t *testing.T) {
	tempDir := t.TempDir()
	pythonExe := filepath.Join(tempDir, "python.exe")
	_ = os.WriteFile(pythonExe, []byte(""), 0755)

	oldExecutor := utils.Executor
	defer func() { utils.Executor = oldExecutor }()

	utils.Executor = &testutil.MockExecutor{
		Output: "Python 3.10.0",
	}

	version, err := GetPythonVersion(tempDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if version != "Python 3.10.0" {
		t.Errorf("Expected Python 3.10.0, got %s", version)
	}

	// Test error case
	_ = os.Remove(pythonExe)
	_, err = GetPythonVersion(tempDir)
	if err == nil {
		t.Error("Expected error for missing python.exe, got nil")
	}
}
