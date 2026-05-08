package heidisql

import (
	_ "embed"
)

//go:embed heidisql.svg
var iconSVG string

func DetectVersions() (string, string) {
	version := "12.7"
	url := "https://www.heidisql.com/downloads/releases/HeidiSQL_12.7_64_Portable.zip"
	return version, url
}

func GetIcon() string {
	return iconSVG
}
