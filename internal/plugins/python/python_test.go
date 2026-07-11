package python

import (
	"fmt"
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"testing"
)

func TestHelperProcess(t *testing.T) {
	testutil.HelperProcess(t)
}

func TestDetectVersions(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	mockJSON := `{"versions": ["3.10.1", "3.10.2", "3.11.0", "3.12.5"]}`
	utils.Client = &testutil.MockHTTPClient{Content: mockJSON}

	versions, urlMap := DetectVersions()
	if len(versions) < 3 {
		t.Errorf("Expected at least 3 versions, got %v", versions)
	}

	if versions[0] != "3.12.5" {
		t.Errorf("Expected 3.12.5, got %s", versions[0])
	}

	if _, ok := urlMap["3.12.5"]; !ok {
		t.Error("Expected URL for 3.12.5")
	}
}

func TestDetectVersions_Error(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	utils.Client = &testutil.MockHTTPClient{Content: ""}

	versions, _ := DetectVersions()
	if len(versions) != 1 || versions[0] != "3.13.13" {
		t.Errorf("Expected fallback version 3.13.13, got %v", versions)
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}

func TestModules(t *testing.T) {
	if GetModules() != nil {
		t.Error("Expected nil modules")
	}
	if GetModuleVersion("test", "path") != "" {
		t.Error("Expected empty module version")
	}
}

func TestGetInfo(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	tmpDir := t.TempDir()
	pythonExe := filepath.Join(tmpDir, "python.exe")
	os.WriteFile(pythonExe, []byte(""), 0755)

	t.Run("Success", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{Output: "pip 22.3.1 from ..."}
		info := GetInfo(tmpDir)
		if info != "Pip 22.3.1" {
			t.Errorf("Expected Pip 22.3.1, got %s", info)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{Err: fmt.Errorf("error")}
		info := GetInfo(tmpDir)
		if info != "" {
			t.Errorf("Expected empty info on error, got %s", info)
		}
	})

	t.Run("NoExe", func(t *testing.T) {
		info := GetInfo("/invalid/path")
		if info != "" {
			t.Errorf("Expected empty info for no exe, got %s", info)
		}
	})
}

func TestUninstallModule(t *testing.T) {
	err := UninstallModule("any", "any")
	if err == nil {
		t.Error("Expected error from UninstallModule")
	}
}

func TestInstallModule(t *testing.T) {
	err := InstallModule(nil, nil, "any", "any", nil)
	if err == nil {
		t.Error("Expected error from InstallModule")
	}
}
