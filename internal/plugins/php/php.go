package php

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

func DetectVersions() ([]string, map[string]string) {
	baseURL := "https://windows.php.net/downloads/releases/archives/"
	content := fetchContent(baseURL)
	arch := utils.GetSystemArch()
	re := regexp.MustCompile(`php-(\d+\.\d+\.\d+)-Win32-vs16-` + arch + `\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	// Map untuk menyimpan patch terbaru per minor version (misal: "8.2" -> 14)
	latestPatches := make(map[string]int)
	seen := make(map[string]bool)

	for _, m := range matches {
		v := m[1]
		if seen[v] { continue }
		seen[v] = true

		var major, minor, patch int
		fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)

		// Batasan: PHP 8.2 ke atas
		if major > 8 || (major == 8 && minor >= 2) {
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
		versions = append(versions, fullVer)
		urlMap[fullVer] = fmt.Sprintf("%sphp-%s-Win32-vs16-%s.zip", baseURL, fullVer, arch)
	}

	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "8.2.12"
		return []string{v}, map[string]string{v: fmt.Sprintf("%sphp-%s-Win32-vs16-%s.zip", baseURL, v, arch)}
	}
	return versions, urlMap
}

func compareVersions(v1, v2 string) int {
	var a1, b1, c1 int
	var a2, b2, c2 int
	fmt.Sscanf(v1, "%d.%d.%d", &a1, &b1, &c1)
	fmt.Sscanf(v2, "%d.%d.%d", &a2, &b2, &c2)
	if a1 != a2 { return a1 - a2 }
	if b1 != b2 { return b1 - b2 }
	return c1 - c2
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "php", "php.svg"))
	return string(data)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
