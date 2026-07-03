package mysql

import (
	"testing"
)

func TestDetectVersions(t *testing.T) {
	versions, urlMap := DetectVersions()
	if len(versions) == 0 {
		t.Error("Expected versions")
	}
	for _, v := range versions {
		if _, ok := urlMap[v]; !ok {
			t.Errorf("Expected URL for version %s", v)
		}
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}
