package nodejs

import (
	_ "embed"
	"fmt"
	"ostenia/internal/plugins/utils"
	"regexp"
	"runtime"
	"sort"
)

//go:embed node.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	arch := "win-x64"
	if runtime.GOARCH == "386" {
		arch = "win-x86"
	}

	content := utils.FetchContent("https://nodejs.org/dist/")
	// Mencocokkan versi v22.x.x dan v24.x.x
	re := regexp.MustCompile(`v(22|24)\.(\d+)\.(\d+)/`)
	matches := re.FindAllStringSubmatch(content, -1)

	// Map untuk menyimpan hanya versi absolut terbaru untuk setiap Major (22 dan 24)
	latestForMajor := make(map[string]string)

	for _, m := range matches {
		major := m[1]
		currentFull := m[1] + "." + m[2] + "." + m[3]

		if existingFull, ok := latestForMajor[major]; !ok || utils.CompareVersions(currentFull, existingFull) > 0 {
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
	sort.Slice(versions, func(i, j int) bool { return utils.CompareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "22.12.0"
		return []string{v}, map[string]string{v: "https://nodejs.org/dist/v22.12.0/node-v22.12.0-win-x64.zip"}
	}
	return versions, urlMap
}

func GetIcon() string {
	return iconSVG
}
