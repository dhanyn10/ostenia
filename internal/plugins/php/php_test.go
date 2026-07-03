package php

import (
	"bytes"
	"io"
	"net/http"
	"ostenia/internal/plugins/utils"
	"testing"
)

type mockHTTPClient struct {
	utils.HTTPClient
	content string
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(m.content)),
	}, nil
}

func TestDetectVersions(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	arch := utils.GetSystemArch()
	mockHTML := `
		<a href="php-8.2.1-Win32-vs16-` + arch + `.zip">php-8.2.1-Win32-vs16-` + arch + `.zip</a>
		<a href="php-8.2.2-Win32-vs16-` + arch + `.zip">php-8.2.2-Win32-vs16-` + arch + `.zip</a>
		<a href="php-8.4.1-Win32-vs16-` + arch + `.zip">php-8.4.1-Win32-vs16-` + arch + `.zip</a>
	`
	utils.Client = &mockHTTPClient{content: mockHTML}

	versions, urlMap := DetectVersions()
	if len(versions) < 2 {
		t.Errorf("Expected at least 2 versions, got %v", versions)
	}

	// Should pick latest patch for each minor
	found84 := false
	found822 := false
	found821 := false
	for _, v := range versions {
		if v == "8.4.1" { found84 = true }
		if v == "8.2.2" { found822 = true }
		if v == "8.2.1" { found821 = true }
	}
	if !found84 || !found822 || found821 {
		t.Errorf("Unexpected versions: found84=%v, found822=%v, found821=%v", found84, found822, found821)
	}

	if _, ok := urlMap["8.2.2"]; !ok {
		t.Error("Expected URL for 8.2.2")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}

func TestModules(t *testing.T) {
	mods := GetModules()
	if len(mods) == 0 {
		t.Error("Expected modules")
	}

	tempDir := t.TempDir()
	phpPath := tempDir

	// Test GetModuleVersion when not installed
	if GetModuleVersion("Composer", phpPath) != "" {
		t.Error("Expected empty version for missing Composer")
	}

	// Test UninstallModule
	err := UninstallModule("Composer", phpPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

    _ = UninstallModule("Unknown", phpPath)
}
