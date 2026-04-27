package heidisql

import (
	"os"
	"path/filepath"
)

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "heidisql", "heidisql.svg"))
	return string(data)
}
