package mysql

import (
	_ "embed"
	"strings"
)

//go:embed mysql.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	// MySQL versions usually fixed for LTS
	versions := []string{"8.4.0", "9.1.0"} // 8.4 LTS and 9.1 (Innovation/LTS equivalent)
	urlMap := make(map[string]string)

	// Base URL for MySQL community server zips
	for _, v := range versions {
		urlMap[v] = "https://dev.mysql.com/get/Downloads/MySQL-" + strings.Split(v, ".")[0] + "." + strings.Split(v, ".")[1] + "/mysql-" + v + "-winx64.zip"
	}

	return versions, urlMap
}

func GetIcon() string {
	return iconSVG
}
