package nginx

import (
	"os"
	"path/filepath"
)

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "nginx", "nginx.svg"))
	return string(data)
}
