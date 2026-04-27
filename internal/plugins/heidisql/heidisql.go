package heidisql

import (
	"os"
	"path/filepath"
)

func DetectVersions() (string, string) {
	version := "12.7"
	url := "https://www.heidisql.com/downloads/releases/HeidiSQL_12.7_64_Portable.zip"
	return version, url
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "heidisql", "heidisql.svg"))
	return string(data)
}
