package nginx

import (
	"fmt"
	"os"
	"path/filepath"
)

func DetectVersions() (string, string) {
	version := "1.24.0"
	url := fmt.Sprintf("https://nginx.org/download/nginx-%s.zip", version)
	return version, url
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "nginx", "nginx.svg"))
	return string(data)
}
