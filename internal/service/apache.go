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

func UpdateApacheConfig(apachePath string, phpDllPath string, phpIniDir string, vhostsContent string, port int, wwwRoot string) error {
	confPath := filepath.Join(apachePath, "conf", "httpd.conf")

	input, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}

	content := string(input)
	
	// 1. Normalisasi Path Ostenia
	absApachePath, _ := filepath.Abs(apachePath)
	normalizedApachePath := strings.ReplaceAll(absApachePath, "\\", "/")
	normalizedWWWRoot := strings.ReplaceAll(wwwRoot, "\\", "/")

	// 2. Aktifkan Module Penting
	reRewrite := regexp.MustCompile(`(?m)^#\s*LoadModule rewrite_module`)
	content = reRewrite.ReplaceAllString(content, "LoadModule rewrite_module")
	reAlias := regexp.MustCompile(`(?m)^#\s*LoadModule alias_module`)
	content = reAlias.ReplaceAllString(content, "LoadModule alias_module")

	// 3. Bersihkan path absolut luar (Laragon dll)
	reAnyAbsPath := regexp.MustCompile(`[A-Za-z]:/[^" \n\t]+`)
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

	// 4. Bersihkan blok-blok definisi lama
	reSRVROOT := regexp.MustCompile(`(?m)^Define\s+SRVROOT\s+.*?\n`)
	reSRV := regexp.MustCompile(`(?m)^ServerRoot\s+.*?\n`)
	reDocRoot := regexp.MustCompile(`(?m)^DocumentRoot\s+.*?\n`)
	reServerName := regexp.MustCompile(`(?m)^ServerName\s+.*?\n`)
	reDocDir := regexp.MustCompile(`(?s)<Directory\s+".*?">\s*#\s*MainDocRoot\n.*?</Directory>\n?`)
	rePHPBlock := regexp.MustCompile(`(?s)# Ostenia PHP Configuration.*?PHPIniDir ".*?"\n?`)
	reFallback := regexp.MustCompile(`(?s)# Ostenia Fallback Configuration.*?\n?# End Ostenia Fallback\n?`)
	reDirIndex := regexp.MustCompile(`(?m)^DirectoryIndex\s+.*?\n`)

	content = reSRVROOT.ReplaceAllString(content, "")
	content = reSRV.ReplaceAllString(content, "")
	content = reDocRoot.ReplaceAllString(content, "")
	content = reServerName.ReplaceAllString(content, "")
	content = reDocDir.ReplaceAllString(content, "")
	content = rePHPBlock.ReplaceAllString(content, "")
	content = reFallback.ReplaceAllString(content, "")
	content = reDirIndex.ReplaceAllString(content, "")

	// 5. Set Header Baru
	header := fmt.Sprintf("Define SRVROOT \"%s\"\nServerRoot \"${SRVROOT}\"\nServerName localhost\n", normalizedApachePath)
	header += fmt.Sprintf("DocumentRoot \"%s\"\n", normalizedWWWRoot)
	header += fmt.Sprintf("<Directory \"%s\"> # MainDocRoot\n    Options Indexes FollowSymLinks\n    AllowOverride All\n    Require all granted\n</Directory>\n", normalizedWWWRoot)

	// Tambahkan Logika Fallback menggunakan Alias dan DirectoryIndex
	header += `
# Ostenia Fallback Configuration
Alias /__ostenia_default "${SRVROOT}/htdocs"
<Directory "${SRVROOT}/htdocs">
    AllowOverride None
    Require all granted
</Directory>

# Urutan pencarian: index.php -> index.html -> fallback ke htdocs bawaan
DirectoryIndex index.php index.html /__ostenia_default/index.html
# End Ostenia Fallback
`

	// Update Port Listen
	rePort := regexp.MustCompile(`(?m)^Listen\s+\d+`)
	if rePort.MatchString(content) {
		content = rePort.ReplaceAllString(content, fmt.Sprintf(`Listen %d`, port))
	}

	// 6. Siapkan Blok PHP
	phpConfigBlock := ""
	if phpDllPath != "" {
		normalizedPhpDll := strings.ReplaceAll(phpDllPath, "\\", "/")
		normalizedPhpIni := strings.ReplaceAll(phpIniDir, "\\", "/")
		phpConfigBlock = fmt.Sprintf(`
# Ostenia PHP Configuration
LoadModule php_module "%s"
AddHandler application/x-httpd-php .php
PHPIniDir "%s"
`, normalizedPhpDll, normalizedPhpIni)
	}

	// 7. Gabungkan Kembali
	finalContent := header + strings.TrimSpace(content) + "\n\n" + phpConfigBlock

	if !strings.Contains(finalContent, "Include conf/extra/httpd-vhosts.conf") {
		finalContent += "\nInclude conf/extra/httpd-vhosts.conf"
	}

	vhostsPath := filepath.Join(apachePath, "conf", "extra", "httpd-vhosts.conf")
	os.MkdirAll(filepath.Dir(vhostsPath), 0755)
	os.WriteFile(vhostsPath, []byte(vhostsContent), 0644)

	return os.WriteFile(confPath, []byte(finalContent), 0644)
}
