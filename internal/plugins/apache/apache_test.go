package apache

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
		<a href="binaries/httpd-2.4.58-win64-VS17.zip">httpd-2.4.58-win64-VS17.zip</a>
		<a href="binaries/httpd-2.4.57-win64-VS17.zip">httpd-2.4.57-win64-VS17.zip</a>
	`
	utils.Client = &mockHTTPClient{content: mockHTML}

	versions, urlMap := DetectVersions()

	if len(versions) == 0 {
		t.Error("Expected versions to be detected")
	}

	found := false
	for _, v := range versions {
		if v == "2.4.66-260223" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected version 2.4.66-260223 to be detected")
	}

	if len(versions) > 0 {
        t.Logf("Detected versions: %v", versions)
        t.Logf("URL map: %v", urlMap)
    }
}

func TestGetIcon(t *testing.T) {
	icon := GetIcon()
	if icon == "" {
		t.Error("Expected non-empty icon SVG")
	}
}
