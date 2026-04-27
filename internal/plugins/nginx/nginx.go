package nginx

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

func DetectVersions() ([]string, map[string]string) {
	content := fetchContent("https://nginx.org/download/")
	// Match versions like nginx-1.29.x.zip
	re := regexp.MustCompile(`nginx-(\d+\.\d+\.\d+)\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)

	for _, m := range matches {
		v := m[1]
		var major, minor int
		fmt.Sscanf(v, "%d.%d", &major, &minor)

		// Batasan: Nginx 1.27 ke atas (Karena 1.29 mungkin belum rilis, saya ambil yang terbaru stabil)
		if (major > 1 || (major == 1 && minor >= 27)) && !seen[v] {
			versions = append(versions, v)
			urlMap[v] = fmt.Sprintf("https://nginx.org/download/nginx-%s.zip", v)
			seen[v] = true
		}
	}

	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "1.27.2"
		return []string{v}, map[string]string{v: fmt.Sprintf("https://nginx.org/download/nginx-%s.zip", v)}
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
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "nginx", "nginx.svg"))
	return string(data)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
