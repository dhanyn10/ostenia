package php

import (
	"testing"
)

func TestDetectVersions(t *testing.T) {
	versions, urlMap := DetectVersions()
	if len(versions) == 0 {
		t.Error("Expected at least one version")
	}
	for _, v := range versions {
		if urlMap[v] == "" {
			t.Errorf("No URL mapping for version %s", v)
		}
	}
}

func TestGetIcon(t *testing.T) {
	icon := GetIcon()
	if icon == "" {
		t.Error("Expected non-empty icon string")
	}
}

func TestGetModules(t *testing.T) {
	modules := GetModules()
	if len(modules) == 0 {
		t.Error("Expected at least one module")
	}
	if modules[0].Name != "Composer" {
		t.Errorf("Expected Composer module, got %s", modules[0].Name)
	}
}
