package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func UpdateApacheConfig(apachePath string, phpDllPath string, phpIniDir string, vhostsContent string, port int) error {
	confPath := filepath.Join(apachePath, "conf", "httpd.conf")

	input, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}

	content := string(input)
	
	// 1. Deteksi Root Baru secara Absolut
	absApachePath, _ := filepath.Abs(apachePath)
	normalizedApachePath := strings.ReplaceAll(absApachePath, "\\", "/")

	// 2. Pembersihan Total Path Kaku
	// Kita hapus baris Define SRVROOT, ServerRoot, dan blok PHP lama agar tidak duplikat/berantakan
	reSRVROOT := regexp.MustCompile(`(?m)^Define\s+SRVROOT\s+.*?\n`)
	reSRV := regexp.MustCompile(`(?m)^ServerRoot\s+.*?\n`)
	rePHPBlock := regexp.MustCompile(`(?s)# Ostenia PHP Configuration.*?PHPIniDir ".*?"\n?`)
	
	content = reSRVROOT.ReplaceAllString(content, "")
	content = reSRV.ReplaceAllString(content, "")
	content = rePHPBlock.ReplaceAllString(content, "")

	// 3. Re-Rooting: Gunakan ${SRVROOT} untuk semua path internal Apache
	// Biasanya Apache bawaan punya path "c:/Apache24" atau semacamnya
	// Kita ganti semua referensi path lama ke variabel ${SRVROOT}
	// Ini membuat config menjadi benar-benar portable
	reOldPath := regexp.MustCompile(`(?i)c:/Apache2[4-9][^ \n\t"]*`)
	content = reOldPath.ReplaceAllString(content, "${SRVROOT}")

	// 4. Set Header Baru (Root & Port)
	header := fmt.Sprintf("Define SRVROOT \"%s\"\nServerRoot \"${SRVROOT}\"\n", normalizedApachePath)
	
	// Update Port Listen
	rePort := regexp.MustCompile(`(?m)^Listen\s+\d+`)
	if rePort.MatchString(content) {
		content = rePort.ReplaceAllString(content, fmt.Sprintf(`Listen %d`, port))
	}

	// 5. Siapkan Blok PHP
	normalizedPhpDll := strings.ReplaceAll(phpDllPath, "\\", "/")
	normalizedPhpIni := strings.ReplaceAll(phpIniDir, "\\", "/")
	
	phpConfigBlock := fmt.Sprintf(`
# Ostenia PHP Configuration
LoadModule php_module "%s"
AddHandler application/x-httpd-php .php
PHPIniDir "%s"
`, normalizedPhpDll, normalizedPhpIni)

	// 6. Gabungkan Kembali
	// Header di paling atas
	// PHP di paling bawah (agar mod_mime sudah terload dan AddHandler tidak error)
	finalContent := header + strings.TrimSpace(content) + "\n\n" + phpConfigBlock

	// 7. Pastikan VHosts disertakan
	if !strings.Contains(finalContent, "Include conf/extra/httpd-vhosts.conf") {
		finalContent += "\nInclude conf/extra/httpd-vhosts.conf"
	}

	// 8. Tulis VHosts
	vhostsContent = strings.ReplaceAll(vhostsContent, "*:80", fmt.Sprintf("*:%d", port))
	vhostsPath := filepath.Join(apachePath, "conf", "extra", "httpd-vhosts.conf")
	os.MkdirAll(filepath.Dir(vhostsPath), 0755)
	os.WriteFile(vhostsPath, []byte(vhostsContent), 0644)

	return os.WriteFile(confPath, []byte(finalContent), 0644)
}
