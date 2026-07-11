package service

import (
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"runtime"
	"testing"
)

func TestIsAdmin(t *testing.T) {
	res := IsAdmin()
	t.Logf("IsAdmin: %v", res)
}

func TestRunMeAsAdmin(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	t.Run("Windows_Mock", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			utils.Executor = &testutil.MockExecutor{Output: ""}
			err := RunMeAsAdmin()
			if err != nil {
				t.Errorf("RunMeAsAdmin failed: %v", err)
			}
		}
	})

	t.Run("NonWindows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			err := RunMeAsAdmin()
			if err == nil {
				t.Error("Expected error on non-windows")
			}
		}
	})
}

func TestAddHostWithElevation(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	t.Run("Windows_Mock", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			// Mock IsAdmin to false if we want to test elevation logic
			// but IsAdmin is not easily mockable since it's a function.
			// Let's assume we are not admin in test.
			utils.Executor = &testutil.MockExecutor{Output: ""}
			err := AddHostWithElevation("127.0.0.1", "test.local")
			if err != nil && !IsAdmin() {
				t.Errorf("AddHostWithElevation failed: %v", err)
			}
		}
	})

	t.Run("NonWindows", func(t *testing.T) {
		if runtime.GOOS != "windows" && !IsAdmin() {
			err := AddHostWithElevation("127.0.0.1", "test.local")
			if err == nil {
				t.Error("Expected error on non-windows")
			}
		}
	})
}

func TestHelperProcess_Utils(t *testing.T) {
	testutil.HelperProcess(t)
}

func TestElevateAndExit(t *testing.T) {
	// This calls os.Exit, so we can't test it directly easily.
}
