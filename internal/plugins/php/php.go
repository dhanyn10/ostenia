package php

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func DetectVersions() ([]string, string) {
	baseURL := "https://windows.php.net/downloads/releases/archives/"
	content := fetchContent(baseURL)
	re := regexp.MustCompile(`php-(\d+\.\d+\.\d+)-Win32-vs16-x64\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			seen[v] = true
		}
	}

	// Sort newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		return []string{"8.2.12"}, baseURL
	}
	return versions, baseURL
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
