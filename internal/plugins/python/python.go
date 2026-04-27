package python

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

func DetectVersions() ([]string, map[string]string) {
	arch := "amd64"
	if runtime.GOARCH == "386" { arch = "win32" }

	content := fetchContent("https://www.python.org/ftp/python/")
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)/`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = "https://www.python.org/ftp/python/" + v + "/python-" + v + "-embed-" + arch + ".zip"
			seen[v] = true
		}
	}

	// Sort newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "3.12.2"
		return []string{v}, map[string]string{v: "https://www.python.org/ftp/python/3.12.2/python-3.12.2-embed-amd64.zip"}
	}
	return versions, urlMap
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "python", "python.svg"))
	return string(data)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
