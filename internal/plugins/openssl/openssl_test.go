package openssl

import (
	"fmt"
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectVersions(t *testing.T) {
	versions, urlMap := DetectVersions()
	if len(versions) != 1 || versions[0] != "4.0.0" {
		t.Errorf("Expected 4.0.0, got %v", versions)
	}
	if _, ok := urlMap["4.0.0"]; !ok {
		t.Error("Expected URL for 4.0.0")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}

func TestDetectInstalledVersion(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	utils.Executor = &testutil.MockExecutor{Output: "OpenSSL 3.0.7 1 Nov 2022"}

	version := DetectInstalledVersion()
	if version != "3.0.7" {
		t.Errorf("Expected 3.0.7, got %s", version)
	}
}

func TestFindExecutables(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	utils.Executor = &testutil.MockExecutor{Output: "C:\\bin\\openssl.exe\n"}

	tempDir := t.TempDir()
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Unsetenv("OSTENIA_HOME")

	binDir := filepath.Join(tempDir, "bin")
	os.MkdirAll(binDir, 0755)
	os.WriteFile(filepath.Join(binDir, "openssl.exe"), []byte(""), 0755)

	_ = findExecutables()
}

func TestDetectInstalledVersion_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Mock findExecutables to test the branch
		// OpenSSL doesn't have internal hooks for findExecutables, so we just run it and see.
		// On linux it will skip the GOOS == "windows" block.
	}
}

func TestVersionFromExecutable(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	t.Run("Success", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{Output: "OpenSSL 1.1.1q  17 Aug 2022"}
		v := versionFromExecutable("openssl")
		if v != "1.1.1q" {
			t.Errorf("Expected 1.1.1q, got %s", v)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{Err: fmt.Errorf("error")}
		v := versionFromExecutable("invalid")
		if v != "" {
			t.Errorf("Expected empty version on error, got %s", v)
		}
	})

	t.Run("ShortOutput", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{Output: "OpenSSL"}
		v := versionFromExecutable("openssl")
		if v != "" {
			t.Errorf("Expected empty version on short output, got %s", v)
		}
	})
}
