package service

import (
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"testing"
)

import (
	"os"
	"path/filepath"
)

func TestNodeAndPythonVersion(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	tempDir := t.TempDir()

	t.Run("GetNodeVersion", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{}
		os.WriteFile(filepath.Join(tempDir, "node.exe"), []byte("#!/bin/sh\necho v1.0.0"), 0755)
		_, err := GetNodeVersion(tempDir)
		if err != nil {
			t.Errorf("GetNodeVersion failed: %v", err)
		}
	})

	t.Run("GetPythonVersion", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{}
		os.WriteFile(filepath.Join(tempDir, "python.exe"), []byte("#!/bin/sh\necho 3.12.0"), 0755)
		_, err := GetPythonVersion(tempDir)
		if err != nil {
			t.Errorf("GetPythonVersion failed: %v", err)
		}
	})

	t.Run("GetPHPVersion", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{}
		os.WriteFile(filepath.Join(tempDir, "php.exe"), []byte("#!/bin/sh\necho 8.3.0"), 0755)
		_, err := GetPHPVersion(tempDir)
		if err != nil {
			t.Errorf("GetPHPVersion failed: %v", err)
		}
	})
}
