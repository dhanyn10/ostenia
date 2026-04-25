package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func GenerateVHost(projectName string, projectPath string, port int) string {
	return fmt.Sprintf(`
<VirtualHost *:%d>
    DocumentRoot "%s"
    ServerName %s.test
    ServerAlias *.%s.test
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>
`, port, projectPath, projectName, projectName, projectPath)
}

func UpdateApacheConfig(apachePath string, phpDllPath string, phpIniDir string, vhostsContent string, port int, wwwRoot string, phpPort int) error {
	confPath := filepath.Join(apachePath, "conf", "httpd.conf")
	phpConfPath := filepath.Join(apachePath, "conf", "extra", "httpd-ostenia-php.conf")

	input, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}

	content := string(input)
	
	// 1. Normalize Paths
	absApachePath, _ := filepath.Abs(apachePath)
	normalizedApachePath := strings.ReplaceAll(absApachePath, "\\", "/")
	normalizedWWWRoot := strings.ReplaceAll(wwwRoot, "\\", "/")

	// 2. Enable Required Modules
	modules := []string{"rewrite_module", "alias_module", "proxy_module", "proxy_fcgi_module"}
	for _, mod := range modules {
		re := regexp.MustCompile(`(?m)^#\s*LoadModule\s+` + mod)
		content = re.ReplaceAllString(content, "LoadModule "+mod)
	}

	// 3. Clean all absolute paths that point to an external Apache folder (like Laragon)
	reAnyAbsPath := regexp.MustCompile(`[A-Za-z]:/[^" \n\t\r]+`)
	content = reAnyAbsPath.ReplaceAllStringFunc(content, func(match string) string {
		if strings.Contains(match, "/bin/apache/") && !strings.Contains(match, normalizedApachePath) {
			parts := strings.Split(match, "/modules/")
			if len(parts) > 1 {
				return "${SRVROOT}/modules/" + parts[1]
			}
			return "${SRVROOT}"
		}
		return match
	})

	// 4. Remove ALL existing Ostenia blocks and common hardcoded directives
	patterns := []string{
		`(?s)# Ostenia System Config.*?# End Ostenia System Config\n?`, // Main config block
		`(?s)# Ostenia PHP Configuration.*?# End Ostenia PHP\n?`,       // Old PHP block
		`(?m)^Define\s+SRVROOT\s+.*?\n`,
		`(?m)^ServerRoot\s+.*?\n`,
		`(?m)^DocumentRoot\s+.*?\n`,
		`(?m)^ServerName\s+.*?\n`,
		`(?m)^DirectoryIndex\s+.*?\n`,
		`(?s)<Directory\s+".*?">\s*#\s*MainDocRoot.*?</Directory>.*?\n`,
		`(?m)^Include\s+conf/extra/httpd-ostenia-php.conf\n?`, // Remove old include
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p)
		content = re.ReplaceAllString(content, "")
	}

	// 5. Build New Main Configuration Header
	header := "# Ostenia System Config\n"
	header += fmt.Sprintf("Define SRVROOT \"%s\"\n", normalizedApachePath)
	header += "ServerRoot \"${SRVROOT}\"\n"
	header += "ServerName localhost\n"
	header += fmt.Sprintf("DocumentRoot \"%s\"\n", normalizedWWWRoot)

	// Main Directory Config
	header += fmt.Sprintf("<Directory \"%s\">\n", normalizedWWWRoot)
	header += "    # MainDocRoot\n"
	header += "    Options Indexes FollowSymLinks\n"
	header += "    AllowOverride All\n"
	header += "    Require all granted\n"
	header += "</Directory>\n"

	// Fallback Configuration
	header += `
# Ostenia Fallback Configuration
Alias /__ostenia_default "${SRVROOT}/htdocs"
<Directory "${SRVROOT}/htdocs">
    AllowOverride None
    Require all granted
</Directory>
DirectoryIndex index.php index.html /__ostenia_default/index.html
# End Ostenia System Config
`

	// 6. Update Listen Port
	rePort := regexp.MustCompile(`(?m)^Listen\s+\d+`)
	if !rePort.MatchString(content) {
		content = "Listen 80\n" + content
	} else {
		content = rePort.ReplaceAllString(content, fmt.Sprintf(`Listen %d`, port))
	}

	// 7. Include PHP config file
	content += "\nInclude conf/extra/httpd-ostenia-php.conf\n"

	// 8. Final Combine
	finalContent := header + "\n" + strings.TrimSpace(content)

	if !strings.Contains(finalContent, "Include conf/extra/httpd-vhosts.conf") {
		finalContent += "\nInclude conf/extra/httpd-vhosts.conf"
	}

	// 9. Write vhosts
	vhostsPath := filepath.Join(apachePath, "conf", "extra", "httpd-vhosts.conf")
	os.MkdirAll(filepath.Dir(vhostsPath), 0755)
	os.WriteFile(vhostsPath, []byte(vhostsContent), 0644)

	// 10. Write PHP config to its own file
	phpConfContent := "# Ostenia PHP Configuration\n"
	if phpPort > 0 {
		phpConfContent += fmt.Sprintf("<FilesMatch \\.php$>\n    SetHandler \"proxy:fcgi://127.0.0.1:%d\"\n</FilesMatch>\n", phpPort)
	} else if phpDllPath != "" {
		// Fallback to Module if DLL exists and FastCGI not active
		normalizedPhpDll := strings.ReplaceAll(phpDllPath, "\\", "/")
		normalizedPhpIni := strings.ReplaceAll(phpIniDir, "\\", "/")
		phpConfContent += fmt.Sprintf(`
LoadModule php_module "%s"
AddHandler application/x-httpd-php .php
PHPIniDir "%s"
`, normalizedPhpDll, normalizedPhpIni)
	}
	phpConfContent += "# End Ostenia PHP\n"

	os.MkdirAll(filepath.Dir(phpConfPath), 0755)
	os.WriteFile(phpConfPath, []byte(phpConfContent), 0644)

	return os.WriteFile(confPath, []byte(finalContent), 0644)
}
