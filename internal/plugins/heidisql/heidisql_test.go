package heidisql

import (
	"testing"
)

func TestDetectVersions(t *testing.T) {
	versions, urlMap := DetectVersions()
	if len(versions) == 0 {
		t.Error("Expected at least one version")
	}
	if versions[0] != "12.17" {
		t.Errorf("Expected version 12.17, got %s", versions[0])
	}
	if urlMap[versions[0]] == "" {
		t.Errorf("No URL mapping for version %s", versions[0])
	}
}

func TestGetIcon(t *testing.T) {
	icon := GetIcon()
	if icon == "" {
		t.Error("Expected non-empty icon string")
	}
}
