package php

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
)

func DetectVersions() ([]string, map[string]string) {
	baseURL := "https://windows.php.net/downloads/releases/archives/"
	content := fetchContent(baseURL)
	arch := utils.GetSystemArch()

	// Match both x64 and x86 versions
	re := regexp.MustCompile(`php-(\d+\.\d+\.\d+)-Win32-vs16-` + arch + `\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)

	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = baseURL + m[0]
			seen[v] = true
		}
	}

	if len(versions) == 0 {
		v := "8.2.12"
		versions = []string{v}
		urlMap[v] = fmt.Sprintf("%sphp-%s-Win32-vs16-%s.zip", baseURL, v, arch)
	}

	return versions, urlMap
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
