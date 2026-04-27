package nodejs

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

func DetectVersions() ([]string, map[string]string) {
	arch := "win-x64"
	if runtime.GOARCH == "386" { arch = "win-x86" }

	content := fetchContent("https://nodejs.org/dist/")
	re := regexp.MustCompile(`v(\d+\.\d+\.\d+)/`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = "https://nodejs.org/dist/v" + v + "/node-v" + v + "-" + arch + ".zip"
			seen[v] = true
		}
	}

	// Sort newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "20.11.1"
		return []string{v}, map[string]string{v: "https://nodejs.org/dist/v20.11.1/node-v20.11.1-win-x64.zip"}
	}
	return versions, urlMap
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "nodejs", "node.svg"))
	return string(data)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
