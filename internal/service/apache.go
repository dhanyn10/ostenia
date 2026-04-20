package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateVHost(projectName string, projectPath string) string {
	return fmt.Sprintf(`
<VirtualHost *:80>
    DocumentRoot "%s"
    ServerName %s.test
    ServerAlias *.%s.test
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>
`, projectPath, projectName, projectName, projectPath)
}

func UpdateApacheConfig(binPath string, phpDllPath string, vhostsContent string) error {
	confPath := filepath.Join(binPath, "conf", "httpd.conf")

	// This is a simplified version. In a real app, we would use a template
	// or specifically search and replace the PHP module path.

	input, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	newLines := []string{}

	phpModuleFound := false
	for _, line := range lines {
		if strings.Contains(line, "LoadModule php_module") {
			newLines = append(newLines, fmt.Sprintf("LoadModule php_module \"%s\"", phpDllPath))
			phpModuleFound = true
		} else {
			newLines = append(newLines, line)
		}
	}

	if !phpModuleFound {
		newLines = append(newLines, fmt.Sprintf("LoadModule php_module \"%s\"", phpDllPath))
	}

	// Write vhosts to a separate file and include it
	vhostsPath := filepath.Join(binPath, "conf", "extra", "httpd-vhosts.conf")
	err = os.WriteFile(vhostsPath, []byte(vhostsContent), 0644)
	if err != nil {
		return err
	}

	return os.WriteFile(confPath, []byte(strings.Join(newLines, "\n")), 0644)
}
