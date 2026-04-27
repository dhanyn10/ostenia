package python

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
	"sort"
)

// DetectVersions scans the Python FTP server for available versions
// and returns only the latest patch for each minor version starting from 3.9,
// verifying that the embeddable package URL actually exists.
func DetectVersions() ([]string, map[string]string) {
	arch := utils.GetSystemArch()
	if arch == "x64" { arch = "amd64" } else { arch = "win32" }

	content := fetchContent("https://www.python.org/ftp/python/")
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)/`) // Match any X.Y.Z/ directory
	matches := re.FindAllStringSubmatch(content, -1)

	latestPatches := make(map[string]int)
	for _, m := range matches {
		v := m[1]
		var major, minor, patch int
		_, err := fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
		if err != nil { continue }

		// Filter: Only include Python 3.9+
		if major == 3 && minor >= 9 {
			minorKey := fmt.Sprintf("%d.%d", major, minor)
			if patch > latestPatches[minorKey] {
				latestPatches[minorKey] = patch
			}
		}
	}

	var versions []string
	urlMap := make(map[string]string)

	for minorKey, patch := range latestPatches {
		fullVer := fmt.Sprintf("%s.%d", minorKey, patch)
		// Construct the expected URL
		expectedURL := fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-embed-%s.zip", fullVer, fullVer, arch)

		// Perform a HEAD request to verify if the URL actually exists
		if checkURLExists(expectedURL) {
			versions = append(versions, fullVer)
			urlMap[fullVer] = expectedURL
		} else {
			fmt.Printf("[Python Detector] Skipping %s: Embed package not found at %s\n", fullVer, expectedURL)
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	if len(versions) == 0 {
		v := "3.12.2" // Fallback to a known good version
		fallbackURL := fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-embed-%s.zip", v, v, arch)
		if checkURLExists(fallbackURL) {
			return []string{v}, map[string]string{v: fallbackURL}
		}
		return []string{}, map[string]string{} // No versions found, even fallback failed
	}
	return versions, urlMap
}

// checkURLExists performs a HEAD request to verify if a URL returns 200 OK.
func checkURLExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		fmt.Printf("[Python Detector] HEAD request failed for %s: %v\n", url, err)
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// compareVersions returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
func compareVersions(v1, v2 string) int {
	var a1, b1, c1 int
	var a2, b2, c2 int
	fmt.Sscanf(v1, "%d.%d.%d", &a1, &b1, &c1)
	fmt.Sscanf(v2, "%d.%d.%d", &a2, &b2, &c2)

	if a1 != a2 { return compare(a1, a2) }
	if b1 != b2 { return compare(b1, b2) }
	return compare(c1, c2)
}

func compare(i, j int) int {
	if i > j { return 1 }
	if i < j { return -1 }
	return 0
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "python", "python.svg"))
	return string(data)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[Python Detector] Failed to fetch content from %s: %v\n", url, err)
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
