package heidisql

import (
	_ "embed"
)

//go:embed heidisql.svg
var iconSVG string

func DetectVersions() (string, string) {
	version := "12.17"
	url := "https://github.com/HeidiSQL/HeidiSQL/releases/download/12.17/HeidiSQL_12.17.0.7270_Setup.exe"
	return version, url
}

func GetIcon() string {
	return iconSVG
}
