package nginx

import (
	_ "embed"
	"fmt"
	"ostenia/internal/plugins/utils"
	"regexp"
	"sort"
)

//go:embed nginx.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	content := utils.FetchContent("https://nginx.org/download/")
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

	sort.Slice(versions, func(i, j int) bool { return utils.CompareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "1.27.2"
		return []string{v}, map[string]string{v: fmt.Sprintf("https://nginx.org/download/nginx-%s.zip", v)}
	}
	return versions, urlMap
}

func GetIcon() string {
	return iconSVG
}
