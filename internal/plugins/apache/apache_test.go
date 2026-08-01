package apache

import (
	"bytes"
	"io"
	"net/http"
	"ostenia/internal/plugins/utils"
	"runtime"
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
		<a href="binaries/httpd-2.4.66-260223-Win64-VS18.zip">httpd-2.4.66-260223-Win64-VS18.zip</a>
		<a href="binaries/httpd-2.4.57-win64-VS17.zip">httpd-2.4.57-win64-VS17.zip</a>
	`
	utils.Client = &mockHTTPClient{content: mockHTML}

	versions, urlMap := DetectVersions()
	_ = urlMap

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

}

func TestGetIcon(t *testing.T) {
	icon := GetIcon()
	if icon == "" {
		t.Error("Expected non-empty icon SVG")
	}
}

func TestDetectVersions_EmptyAndArchs(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	// 1. Empty content fallback
	utils.Client = &mockHTTPClient{content: ""}
	versions, _ := DetectVersions()
	if len(versions) != 1 || versions[0] != "2.4.66-260223" {
		t.Errorf("Expected fallback version, got %v", versions)
	}

	// 2. Mock 386 architecture
	// Since runtime.GOARCH cannot be directly reassigned, we can test by mocking content with both 32-bit and 64-bit patterns
	// so the compiled regex (which depends on the architecture) matching the HTML can be simulated.
	if runtime.GOARCH == "386" {
		mockHTML32 := `
			<a href="binaries/httpd-2.4.55-win32-vs17.zip">httpd-2.4.55-win32-vs17.zip</a>
		`
		utils.Client = &mockHTTPClient{content: mockHTML32}
		versions32, _ := DetectVersions()
		if len(versions32) == 0 {
			t.Error("Expected 32-bit version to be detected on 386")
		}
	} else {
		mockHTML64 := `
			<a href="binaries/httpd-2.4.55-Win64-VS17.zip">httpd-2.4.55-Win64-VS17.zip</a>
		`
		utils.Client = &mockHTTPClient{content: mockHTML64}
		versions64, _ := DetectVersions()
		if len(versions64) == 0 {
			t.Error("Expected 64-bit version to be detected on x64")
		}
	}
}
