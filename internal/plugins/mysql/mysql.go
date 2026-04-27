package mysql

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func DetectVersions() ([]string, map[string]string) {
	content := fetchContent("https://dev.mysql.com/downloads/mysql/")
	re := regexp.MustCompile(`mysql-(\d+\.\d+\.\d+)-winx64\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	var versions []string
	urlMap := make(map[string]string)
	seen := make(map[string]bool)
	for _, m := range matches {
		v := m[1]
		if !seen[v] {
			versions = append(versions, v)
			urlMap[v] = "https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-" + m[0]
			seen[v] = true
		}
	}

	if len(versions) == 0 {
		v := "8.0.40"
		return []string{v}, map[string]string{v: "https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.40-winx64.zip"}
	}
	return versions, urlMap
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "mysql", "mysql.svg"))
	return string(data)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
