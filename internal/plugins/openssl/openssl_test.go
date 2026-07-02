package openssl

import (
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
