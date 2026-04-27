package openssl

import (
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
)

func DetectVersions() (string, string) {
	version := "4.0.0"
	arch := utils.GetSystemArch()
	var url string
	if arch == "x64" {
		url = "https://slproweb.com/download/Win64OpenSSL_Light-4_0_0.exe"
	} else {
		url = "https://slproweb.com/download/Win32OpenSSL_Light-4_0_0.exe"
	}
	return version, url
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "openssl", "openssl.svg"))
	return string(data)
}
