package service

import (
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

func TestUpdatePaths(t *testing.T) {
	mockSys := NewMockSystem()
	mockSys.PathUser = "C:\\some\\path;C:\\ostenia\\old-php"
	mockSys.PathMachine = "C:\\system\\path"

	t.Run("UpdatePHPPath", func(t *testing.T) {
		err := UpdatePHPPath(mockSys, "C:\\ostenia\\new-php", true)
		if err != nil {
			t.Errorf("UpdatePHPPath failed: %v", err)
		}
	})

	t.Run("UpdateNodePath", func(t *testing.T) {
		err := UpdateNodePath(mockSys, "C:\\ostenia\\node", true)
		if err != nil {
			t.Errorf("UpdateNodePath failed: %v", err)
		}
	})

	t.Run("UpdatePythonPath", func(t *testing.T) {
		err := UpdatePythonPath(mockSys, "C:\\ostenia\\python", true)
		if err != nil {
			t.Errorf("UpdatePythonPath failed: %v", err)
		}
	})

	t.Run("CheckPaths", func(t *testing.T) {
		if !IsPathInUserPath(mockSys, "C:\\some\\path") {
			t.Error("Expected path in user path")
		}
		if IsPathInSystemPath(mockSys, "C:\\other\\path") {
			t.Error("Did not expect other path in system path")
		}
	})
}
