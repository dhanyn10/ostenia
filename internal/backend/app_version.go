package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
)

// Default GitHub repository details
const (
	defaultGithubOwner = "dhanyn10"
	defaultGithubRepo  = "ostenia"
)

// updateBaseURL can be overridden in tests
var updateBaseURL = "https://api.github.com/repos/" + defaultGithubOwner + "/" + defaultGithubRepo

// AppUpdateInfo contains release information retrieved from GitHub.
type AppUpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseURL     string `json:"releaseUrl"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	PublishedAt    string `json:"publishedAt,omitempty"`
}

type githubReleaseResponse struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
}

type githubTagResponse struct {
	Name string `json:"name"`
}

// GetAppVersion returns the current version string of Ostenia dynamically from config or wails.json.
func (a *App) GetAppVersion() string {
	if a != nil && a.cfg != nil && a.cfg.AppVersion != "" {
		return config.FormatVersion(a.cfg.AppVersion)
	}
	return config.GetAppVersion()
}

var updateHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}

// CheckForAppUpdates checks GitHub for latest release or tag and compares with current version.
func (a *App) CheckForAppUpdates() (*AppUpdateInfo, error) {
	currentVer := a.GetAppVersion()
	info := &AppUpdateInfo{
		CurrentVersion: currentVer,
		LatestVersion:  currentVer,
		HasUpdate:      false,
		ReleaseURL:     fmt.Sprintf("https://github.com/%s/%s/releases", defaultGithubOwner, defaultGithubRepo),
	}

	// 1. Try /releases/latest first
	releaseURL := updateBaseURL + "/releases/latest"
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return info, err
	}
	req.Header.Set("User-Agent", "Ostenia-App/"+currentVer)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := updateHTTPClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var release githubReleaseResponse
			if decodeErr := json.NewDecoder(resp.Body).Decode(&release); decodeErr == nil && release.TagName != "" {
				info.LatestVersion = release.TagName
				info.ReleaseURL = release.HTMLURL
				info.ReleaseNotes = release.Body
				info.PublishedAt = release.PublishedAt
				info.HasUpdate = isNewerVersion(info.LatestVersion, info.CurrentVersion)
				return info, nil
			}
		}
	}

	// 2. Fallback to /tags if /releases/latest is not found (e.g. 404) or empty
	tagsURL := updateBaseURL + "/tags"
	tagReq, err := http.NewRequest("GET", tagsURL, nil)
	if err != nil {
		return info, nil
	}
	tagReq.Header.Set("User-Agent", "Ostenia-App/"+currentVer)
	tagReq.Header.Set("Accept", "application/vnd.github.v3+json")

	tagResp, err := updateHTTPClient.Do(tagReq)
	if err != nil {
		return info, nil
	}
	defer tagResp.Body.Close()

	if tagResp.StatusCode == http.StatusOK {
		var tags []githubTagResponse
		body, readErr := io.ReadAll(tagResp.Body)
		if readErr == nil && json.Unmarshal(body, &tags) == nil && len(tags) > 0 {
			info.LatestVersion = tags[0].Name
			info.ReleaseURL = fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", defaultGithubOwner, defaultGithubRepo, tags[0].Name)
			info.HasUpdate = isNewerVersion(info.LatestVersion, info.CurrentVersion)
		}
	}

	return info, nil
}

// isNewerVersion returns true if remote version is strictly greater than local version.
func isNewerVersion(remote, local string) bool {
	rClean := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(remote), "v"), "V")
	lClean := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(local), "v"), "V")
	return utils.CompareVersions(rClean, lClean) > 0
}
