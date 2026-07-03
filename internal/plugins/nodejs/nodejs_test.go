package nodejs

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

	mockHTML := `
		<a href="v22.1.0/">v22.1.0/</a>
		<a href="v22.2.0/">v22.2.0/</a>
		<a href="v24.1.0/">v24.1.0/</a>
	`
	utils.Client = &mockHTTPClient{content: mockHTML}

	versions, urlMap := DetectVersions()
	if len(versions) != 2 {
		t.Errorf("Expected 2 versions (latest of 22 and 24), got %v", versions)
	}

	if versions[0] != "24.1.0" {
		t.Errorf("Expected 24.1.0, got %s", versions[0])
	}
	if versions[1] != "22.2.0" {
		t.Errorf("Expected 22.2.0, got %s", versions[1])
	}

	if _, ok := urlMap["24.1.0"]; !ok {
		t.Error("Expected URL for 24.1.0")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}
