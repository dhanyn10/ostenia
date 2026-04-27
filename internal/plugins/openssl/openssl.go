package openssl

import (
	"os"
	"path/filepath"
)

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "openssl", "openssl.svg"))
	return string(data)
}
