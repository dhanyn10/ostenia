package heidisql

import (
	_ "embed"
)

//go:embed heidisql.svg
var iconSVG string

func DetectVersions() (string, string) {
	version := "12.8"
	url := "https://www.heidisql.com/downloads/releases/HeidiSQL_12.8_Setup.exe"
	return version, url
}

func GetIcon() string {
	return iconSVG
}
