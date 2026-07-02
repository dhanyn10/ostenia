package heidisql

import (
	"testing"
)

func TestDetectVersions(t *testing.T) {
	versions, urlMap := DetectVersions()
	if len(versions) != 1 || versions[0] != "12.17" {
		t.Errorf("Expected 12.17, got %v", versions)
	}
	if _, ok := urlMap["12.17"]; !ok {
		t.Error("Expected URL for 12.17")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}
