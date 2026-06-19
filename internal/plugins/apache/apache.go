package apache

import (
	_ "embed"
	"ostenia/internal/plugins/utils"
	"regexp"
	"runtime"
)

//go:embed apache.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	arch := "x64"
	if runtime.GOARCH == "386" {
		arch = "x86"
	}

	binBase := "https://www.apachelounge.com/download/VS18/binaries/"
	content := utils.FetchContent("https://www.apachelounge.com/download/")

	rePattern := `httpd-(2\.4\.\d+-\d+)-Win64-VS\d+\.zip`
	if arch == "x86" { rePattern = `httpd-(2\.4\.\d+-\d+)-win32-vs\d+\.zip` }
	re := regexp.MustCompile(rePattern)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = binBase + m[0]
			seen[v] = true
		}
	}
	// Sort newest first
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	if len(versions) == 0 {
		v := "2.4.66-260223"
		return []string{v}, map[string]string{v: binBase + "httpd-2.4.66-260223-Win64-VS18.zip"}
	}
	return versions, urlMap
}

func GetIcon() string {
	return iconSVG
}
