package service

import (
	"fmt"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"testing"
)

func TestPathExistsInString(t *testing.T) {
	tests := []struct {
		pathString string
		targetPath string
		want       bool
	}{
		{"C:\\bin;C:\\Windows", "C:\\bin", true},
		{"C:\\bin;C:\\Windows", "C:\\windows", true}, // Case insensitive
		{"C:\\bin;C:\\Windows", "C:\\other", false},
		{"", "C:\\bin", false},
		{"C:\\ostenia\\php-8.1", "C:\\ostenia\\php-8.1", true},
		{"C:\\ostenia\\php-8.1", "C:\\ostenia\\php-8.2", false},
	}

	for _, tt := range tests {
		got := pathExistsInString(tt.pathString, tt.targetPath)
		if got != tt.want {
			t.Errorf("pathExistsInString(%q, %q) = %v, want %v", tt.pathString, tt.targetPath, got, tt.want)
		}
	}
}

func setupEnvTest(t *testing.T) func() {
	origExecutor := utils.Executor
	utils.Executor = &testutil.MockExecutor{Output: ""}

	getPathOverride = func(target string) (string, error) {
		return "C:\\some\\path;C:\\ostenia\\old-php", nil
	}

	return func() {
		utils.Executor = origExecutor
		getPathOverride = nil
	}
}

func TestUpdatePHPPath(t *testing.T) {
	cleanup := setupEnvTest(t)
	defer cleanup()

	err := UpdatePHPPath("C:\\ostenia\\new-php", true)
	if err != nil {
		t.Errorf("UpdatePHPPath failed: %v", err)
	}
}

func TestUpdateNodePath(t *testing.T) {
	cleanup := setupEnvTest(t)
	defer cleanup()

	err := UpdateNodePath("C:\\ostenia\\node", true)
	if err != nil {
		t.Errorf("UpdateNodePath failed: %v", err)
	}
}

func TestUpdatePythonPath(t *testing.T) {
	cleanup := setupEnvTest(t)
	defer cleanup()

	err := UpdatePythonPath("C:\\ostenia\\python", true)
	if err != nil {
		t.Errorf("UpdatePythonPath failed: %v", err)
	}
}

func TestSetPath(t *testing.T) {
	cleanup := setupEnvTest(t)
	defer cleanup()

	err := SetPath("C:\\new\\path", "User")
	if err != nil {
		t.Errorf("SetPath User failed: %v", err)
	}

	// This will trigger elevation code on Windows if not admin
	_ = SetPath("C:\\new\\path", "Machine")
}

func TestNotifyEnvironmentUpdate(t *testing.T) {
	notifyEnvironmentUpdate()
	NotifyEnvironmentUpdate()
}

func TestGetPath_Mocked(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	// 1. Success case
	utils.Executor = &testutil.MockExecutor{Output: "C:\\bin;C:\\Windows"}
	path, err := GetPath("User")
	if err != nil {
		t.Errorf("Expected success, got err: %v", err)
	}
	if path != "C:\\bin;C:\\Windows" {
		t.Errorf("Expected path C:\\bin;C:\\Windows, got %s", path)
	}

	// 2. Error case
	utils.Executor = &testutil.MockExecutor{Err: fmt.Errorf("powershell failed")}
	_, err = GetPath("User")
	if err == nil {
		t.Error("Expected error from failing powershell")
	}
}

func TestSetPath_MockedErrors(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	// 1. User target error
	utils.Executor = &testutil.MockExecutor{Err: fmt.Errorf("powershell failed")}
	err := SetPath("C:\\path", "User")
	if err == nil {
		t.Error("Expected error when SetPath fails")
	}
}

func TestSetPath_MachineNotAdmin(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	// 1. Success case
	utils.Executor = &testutil.MockExecutor{Output: ""}
	err := SetPath("C:\\path", "Machine")
	if err != nil {
		t.Errorf("Expected SetPath Machine to succeed with MockExecutor, but got: %v", err)
	}

	// 2. Error case (powershell elevated command fails)
	utils.Executor = &testutil.MockExecutor{Err: fmt.Errorf("powershell elevated failed")}
	err = SetPath("C:\\path", "Machine")
	if err == nil {
		t.Error("Expected error when elevated command fails")
	}
}

func TestCheckPaths(t *testing.T) {
	cleanup := setupEnvTest(t)
	defer cleanup()

	// Test original implementations
	getPathOverride = func(target string) (string, error) {
		if target == "User" {
			return "C:\\user\\path", nil
		}
		return "C:\\machine\\path", nil
	}

	if !IsPathInUserPath("C:\\user\\path") {
		t.Error("Expected path in user path")
	}
	if !IsPathInSystemPath("C:\\machine\\path") {
		t.Error("Expected path in system path")
	}
	if IsPathInSystemPath("C:\\other\\path") {
		t.Error("Did not expect other path in system path")
	}
}
