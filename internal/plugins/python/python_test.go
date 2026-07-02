package python

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
	if m.content == "" {
		return nil, io.EOF
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(m.content)),
	}, nil
}

func TestDetectVersions(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	mockJSON := `{"versions": ["3.10.1", "3.10.2", "3.11.0", "3.12.5"]}`
	utils.Client = &mockHTTPClient{content: mockJSON}

	versions, urlMap := DetectVersions()
	if len(versions) < 3 {
		t.Errorf("Expected at least 3 versions, got %v", versions)
	}

	if versions[0] != "3.12.5" {
		t.Errorf("Expected 3.12.5, got %s", versions[0])
	}

	if _, ok := urlMap["3.12.5"]; !ok {
		t.Error("Expected URL for 3.12.5")
	}
}

func TestDetectVersions_Error(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	utils.Client = &mockHTTPClient{content: ""}

	versions, _ := DetectVersions()
	if len(versions) != 1 || versions[0] != "3.13.13" {
		t.Errorf("Expected fallback version 3.13.13, got %v", versions)
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}

func TestModules(t *testing.T) {
    if GetModules() != nil {
        t.Error("Expected nil modules")
    }
    if GetModuleVersion("test", "path") != "" {
        t.Error("Expected empty module version")
    }
}
