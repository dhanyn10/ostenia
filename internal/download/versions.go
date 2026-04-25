package download

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
)

// getSystemArch is a helper to determine the system architecture for download URLs.
func getSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

// fetchContent fetches content from a given URL.
func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// parseMatches extracts version strings from content using a regex and sorts them (newest first).
func parseMatches(content string, re *regexp.Regexp) []string {
	matches := re.FindAllStringSubmatch(content, -1)
	var versions []string
	seen := make(map[string]bool)
	for _, m := range matches {
		// Assuming the first submatch group is the version string
		if !seen[m[1]] {
			versions = append(versions, m[1])
			seen[m[1]] = true
		}
	}
	// Reverse to get newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}
	return versions
}

// DetectPHPVersions detects available PHP versions from the official download page.
func DetectPHPVersions() ([]string, string) {
	arch := getSystemArch()
	baseURL := "https://downloads.php.net/~windows/releases/"
	content := fetchContent(baseURL)
	re := regexp.MustCompile(`php-(8\.\d+\.\d+)-Win32-vs16-` + arch + `\.zip`)
	versions := parseMatches(content, re)
	if len(versions) == 0 {
		return []string{"8.3.6"}, baseURL
	}
	return versions, baseURL
}

// DetectApacheVersions detects available Apache versions from Apache Lounge.
func DetectApacheVersions() ([]string, map[string]string) {
	arch := getSystemArch()
	baseURL := "https://www.apachelounge.com/download/"
	binBase := "https://www.apachelounge.com/download/VS18/binaries/"
	content := fetchContent(baseURL)

	rePattern := `httpd-(2\.4\.\d+-\d+)-Win64-VS\d+\.zip`
	if arch == "x86" {
		rePattern = `httpd-(2\.4\.\d+-\d+)-win32-vs\d+\.zip`
	}

	re := regexp.MustCompile(rePattern)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = binBase + m[0] // m[0] is the full matched string (e.g., httpd-2.4.x-Win64-VS18.zip)
			seen[v] = true
		}
	}
	// Reverse to get newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "2.4.66-260223"
		return []string{v}, map[string]string{v: binBase + "httpd-2.4.66-260223-Win64-VS18.zip"}
	}
	return versions, urlMap
}

// DetectMySQLVersions detects available MySQL versions from the archives page.
func DetectMySQLVersions() ([]string, map[string]string) {
	arch := "winx64"
	if getSystemArch() == "x86" {
		arch = "win32"
	}
	content := fetchContent("https://downloads.mysql.com/archives/community/")

	re := regexp.MustCompile(`mysql-(\d+\.\d+\.\d+)-` + arch + `\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = fmt.Sprintf("https://downloads.mysql.com/archives/get/p/23/file/mysql-%s-%s.zip", v, arch)
			seen[v] = true
		}
	}
	// Reverse to get newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "8.0.37"
		return []string{v}, map[string]string{v: fmt.Sprintf("https://downloads.mysql.com/archives/get/p/23/file/mysql-%s-%s.zip", v, arch)}
	}
	return versions, urlMap
}
