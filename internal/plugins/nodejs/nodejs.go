package nodejs

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"sort"
)

//go:embed node.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	arch := "win-x64"
	if runtime.GOARCH == "386" { arch = "win-x86" }

	content := fetchContent("https://nodejs.org/dist/")
	// Mencocokkan versi v22.x.x dan v24.x.x
	re := regexp.MustCompile(`v(22|24)\.(\d+)\.(\d+)/`)
	matches := re.FindAllStringSubmatch(content, -1)

	// Map untuk menyimpan hanya versi absolut terbaru untuk setiap Major (22 dan 24)
	latestForMajor := make(map[string]string)

	for _, m := range matches {
		major := m[1]
		currentFull := m[1] + "." + m[2] + "." + m[3]

		if existingFull, ok := latestForMajor[major]; !ok || compareVersions(currentFull, existingFull) > 0 {
			latestForMajor[major] = currentFull
		}
	}

	var versions []string
	urlMap := make(map[string]string)
	for _, fullVer := range latestForMajor {
		versions = append(versions, fullVer)
		urlMap[fullVer] = fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s-%s.zip", fullVer, fullVer, arch)
	}

	// Urutkan dari yang terbaru (24 dulu baru 22)
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "22.12.0"
		return []string{v}, map[string]string{v: "https://nodejs.org/dist/v22.12.0/node-v22.12.0-win-x64.zip"}
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
	return iconSVG
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
