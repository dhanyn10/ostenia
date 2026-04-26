package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// getSystemArch is a helper to determine the system architecture for download URLs.
func getSystemArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return "x86"
}

// getInstalledVersionPaths checks for installed versions and returns their full paths.
func getInstalledVersionPaths(baseDir string, category string, checkFile string) map[string]string {
	installedPaths := make(map[string]string)
	compDir := filepath.Join(baseDir, "bin", category)
	entries, err := os.ReadDir(compDir)
	if err != nil {
		return installedPaths
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			ver := entry.Name()
			if idx := strings.Index(ver, "-"); idx != -1 {
				ver = ver[idx+1:]
			}

			potentialPaths := []string{
				filepath.Join(compDir, entry.Name(), checkFile),
				filepath.Join(compDir, entry.Name(), "Apache24", checkFile), // Apache specific
				filepath.Join(compDir, entry.Name(), "bin", checkFile),      // OpenSSL specific
			}

			for _, p := range potentialPaths {
				if _, err := os.Stat(p); err == nil {
					installedPaths[ver] = p
					break
				}
			}
		}
	}
	return installedPaths
}

// getOpenSSLVersion executes 'openssl version' and parses the output.
func getOpenSSLVersion(opensslCmd string) string {
	cmd := exec.Command(opensslCmd, "version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`OpenSSL\s+([\d\.]+[a-z]?)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return "Installed"
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

// parseMatches extracts version strings from content using a regex and sorts them.
func parseMatches(content string, re *regexp.Regexp) []string {
	matches := re.FindAllStringSubmatch(content, -1)
	var versions []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m[1]] {
			versions = append(versions, m[1])
			seen[m[1]] = true
		}
	}
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}
	return versions
}

// DetectPHPVersions detects available PHP versions.
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

// DetectApacheVersions detects available Apache versions.
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
			urlMap[v] = binBase + m[0]
			seen[v] = true
		}
	}
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "2.4.66-260223"
		return []string{v}, map[string]string{v: binBase + "httpd-2.4.66-260223-Win64-VS18.zip"}
	}
	return versions, urlMap
}

// DetectMySQLVersions detects available MySQL versions.
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
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "8.0.37"
		return []string{v}, map[string]string{v: fmt.Sprintf("https://downloads.mysql.com/archives/get/p/23/file/mysql-%s-%s.zip", v, arch)}
	}
	return versions, urlMap
}

// DetectNodeVersions detects available Node.js versions based on specific LTS list.
func DetectNodeVersions() ([]string, map[string]string) {
	versions := []string{"24.15.0", "22.22.2", "20.20.2"}
	arch := getSystemArch()

	urlMap := make(map[string]string)
	for _, v := range versions {
		urlMap[v] = fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s-win-%s.zip", v, v, arch)
	}
	return versions, urlMap
}
