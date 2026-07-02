package nginx

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
		<a href="nginx-1.27.1.zip">nginx-1.27.1.zip</a>
		<a href="nginx-1.27.2.zip">nginx-1.27.2.zip</a>
		<a href="nginx-1.25.0.zip">nginx-1.25.0.zip</a>
	`
	utils.Client = &mockHTTPClient{content: mockHTML}

	versions, urlMap := DetectVersions()
	if len(versions) < 2 {
		t.Errorf("Expected at least 2 versions (1.27.1, 1.27.2), got %v", versions)
	}
	if versions[0] != "1.27.2" {
		t.Errorf("Expected latest version to be 1.27.2, got %s", versions[0])
	}
	if _, ok := urlMap["1.27.2"]; !ok {
		t.Error("Expected URL for 1.27.2")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}
