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
