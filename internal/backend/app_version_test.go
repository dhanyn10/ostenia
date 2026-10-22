package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_GetAppVersion(t *testing.T) {
	app := &App{}
	if got := app.GetAppVersion(); got != "v0.10.0" {
		t.Errorf("GetAppVersion() = %v, want %v", got, "v0.10.0")
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		remote string
		local  string
		want   bool
	}{
		{"v0.11.0", "v0.10.0", true},
		{"0.11.0", "0.10.0", true},
		{"v0.10.1", "v0.10.0", true},
		{"v1.0.0", "v0.10.0", true},
		{"v0.10.0", "v0.10.0", false},
		{"v0.9.0", "v0.10.0", false},
		{"0.8.0", "0.10.0", false},
		{"v0.10.0", "v0.11.0", false},
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.remote, tt.local)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.remote, tt.local, got, tt.want)
		}
	}
}

func TestApp_CheckForAppUpdates_FromRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(githubReleaseResponse{
				TagName:     "v0.11.0",
				Name:        "Release 0.11.0",
				HTMLURL:     "https://github.com/dhanyn10/ostenia/releases/tag/v0.11.0",
				Body:        "New features included",
				PublishedAt: "2026-10-22T00:00:00Z",
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldBaseURL := updateBaseURL
	updateBaseURL = server.URL
	defer func() { updateBaseURL = oldBaseURL }()

	app := &App{}
	info, err := app.CheckForAppUpdates()
	if err != nil {
		t.Fatalf("CheckForAppUpdates() error = %v", err)
	}

	if !info.HasUpdate {
		t.Errorf("expected HasUpdate to be true for v0.11.0 against %s", info.CurrentVersion)
	}
	if info.LatestVersion != "v0.11.0" {
		t.Errorf("expected LatestVersion to be v0.11.0, got %s", info.LatestVersion)
	}
	if info.ReleaseURL != "https://github.com/dhanyn10/ostenia/releases/tag/v0.11.0" {
		t.Errorf("unexpected ReleaseURL: %s", info.ReleaseURL)
	}
}

func TestApp_CheckForAppUpdates_FallbackTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/tags" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]githubTagResponse{
				{Name: "v0.12.0"},
				{Name: "v0.10.0"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	oldBaseURL := updateBaseURL
	updateBaseURL = server.URL
	defer func() { updateBaseURL = oldBaseURL }()

	app := &App{}
	info, err := app.CheckForAppUpdates()
	if err != nil {
		t.Fatalf("CheckForAppUpdates() error = %v", err)
	}

	if !info.HasUpdate {
		t.Errorf("expected HasUpdate to be true for fallback tag v0.12.0")
	}
	if info.LatestVersion != "v0.12.0" {
		t.Errorf("expected LatestVersion to be v0.12.0, got %s", info.LatestVersion)
	}
}

func TestApp_CheckForAppUpdates_NetworkError(t *testing.T) {
	oldBaseURL := updateBaseURL
	updateBaseURL = "http://127.0.0.1:0" // Invalid port causes immediate connection error
	defer func() { updateBaseURL = oldBaseURL }()

	app := &App{}
	info, err := app.CheckForAppUpdates()
	if err != nil {
		t.Fatalf("expected graceful return even on connection failure, got err: %v", err)
	}
	if info.HasUpdate {
		t.Errorf("expected HasUpdate to be false on network error")
	}
	if info.CurrentVersion != "v0.10.0" {
		t.Errorf("expected CurrentVersion to be v0.10.0, got %s", info.CurrentVersion)
	}
}
