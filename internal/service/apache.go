package service

import (
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/ssl"
	"path/filepath"
	"regexp"
	"strings"
)

func GenerateVHost(projectName string, projectPath string, port int) string {
	if port <= 0 {
		port = 80
	}
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

func GenerateProxyVHost(projectName string, targetPort int, listenPort int, httpsEnabled bool, sslDir string) string {
	if listenPort <= 0 {
		listenPort = 80
	}

	vhost := fmt.Sprintf(`
<VirtualHost *:%d>
    ServerName %s.test
    ServerAlias *.%s.test

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:%d/
    ProxyPassReverse / http://127.0.0.1:%d/
</VirtualHost>
`, listenPort, projectName, projectName, targetPort, targetPort)

	if httpsEnabled {
		crtPath := strings.ReplaceAll(filepath.Join(sslDir, projectName+".test.crt"), "\\", "/")
		keyPath := strings.ReplaceAll(filepath.Join(sslDir, projectName+".test.key"), "\\", "/")

		vhost += fmt.Sprintf(`
<VirtualHost *:443>
    ServerName %s.test
    ServerAlias *.%s.test

    SSLEngine on
    SSLCertificateFile "%s"
    SSLCertificateKeyFile "%s"

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:%d/
    ProxyPassReverse / http://127.0.0.1:%d/
</VirtualHost>
`, projectName, projectName, crtPath, keyPath, targetPort, targetPort)
	}

	return vhost
}

func UpdateApacheConfig(apachePath string, phpDllPath string, phpIniDir string, vhostsContent string, port int, wwwRoot string, phpPort int, httpsEnabled bool) error {
	confPath := filepath.Join(apachePath, "conf", "httpd.conf")
	phpConfPath := filepath.Join(apachePath, "conf", "extra", "httpd-ostenia-php.conf")
	sslConfPath := filepath.Join(apachePath, "conf", "extra", "httpd-ostenia-ssl.conf")

	input, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}

	content := string(input)
	
	absApachePath, _ := filepath.Abs(apachePath)
	normalizedApachePath := strings.ReplaceAll(absApachePath, "\\", "/")
	normalizedWWWRoot := strings.ReplaceAll(wwwRoot, "\\", "/")

	// Modules
	modules := []string{"rewrite_module", "alias_module", "proxy_module", "proxy_http_module", "proxy_fcgi_module", "socache_shmcb_module"}
	if httpsEnabled {
		modules = append(modules, "ssl_module")
	}
	for _, mod := range modules {
		re := regexp.MustCompile(`(?m)^#\s*LoadModule\s+` + mod)
		content = re.ReplaceAllString(content, "LoadModule "+mod)
	}

	// Remove existing Ostenia blocks
	patterns := []string{
		`(?s)# Ostenia System Config.*?# End Ostenia System Config\n?`,
		`(?s)# Ostenia PHP Configuration.*?# End Ostenia PHP\n?`,
		`(?s)# Ostenia SSL Configuration.*?# End Ostenia SSL\n?`,
		`(?m)^Define\s+SRVROOT\s+.*?\n`,
		`(?m)^ServerRoot\s+.*?\n`,
		`(?m)^DocumentRoot\s+.*?\n`,
		`(?m)^ServerName\s+.*?\n`,
		`(?m)^DirectoryIndex\s+.*?\n`,
		`(?s)<Directory\s+".*?">\s*#\s*MainDocRoot.*?</Directory>.*?\n`,
		`(?m)^Include\s+conf/extra/httpd-ostenia-php.conf\n?`,
		`(?m)^Include\s+conf/extra/httpd-ostenia-ssl.conf\n?`,
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p)
		content = re.ReplaceAllString(content, "")
	}

	// Main Header
	header := "# Ostenia System Config\n"
	header += fmt.Sprintf("Define SRVROOT \"%s\"\n", normalizedApachePath)
	header += "ServerRoot \"${SRVROOT}\"\n"
	header += "ServerName localhost\n"
	header += fmt.Sprintf("DocumentRoot \"%s\"\n", normalizedWWWRoot)
	header += fmt.Sprintf("<Directory \"%s\">\n    Options Indexes FollowSymLinks\n    AllowOverride All\n    Require all granted\n</Directory>\n", normalizedWWWRoot)
	header += "DirectoryIndex index.php index.html\n# End Ostenia System Config\n"

	if port > 0 {
		rePort := regexp.MustCompile(`(?m)^Listen\s+\d+`)
		if !rePort.MatchString(content) {
			content = fmt.Sprintf("Listen %d\n", port) + content
		} else {
			content = rePort.ReplaceAllString(content, fmt.Sprintf(`Listen %d`, port))
		}
	}

	content += "\nInclude conf/extra/httpd-ostenia-php.conf\n"
	if httpsEnabled {
		content += "Include conf/extra/httpd-ostenia-ssl.conf\n"
	}
	if vhostsContent != "" {
		vhostsConfPath := filepath.Join(apachePath, "conf", "extra", "httpd-ostenia-vhosts.conf")
		os.WriteFile(vhostsConfPath, []byte(vhostsContent), 0644)
		if !strings.Contains(content, "Include conf/extra/httpd-ostenia-vhosts.conf") {
			content += "\nInclude conf/extra/httpd-ostenia-vhosts.conf\n"
		}
	}

	finalContent := header + "\n" + strings.TrimSpace(content)

	// PHP Config
	phpConfContent := "# Ostenia PHP Configuration\n"
	if phpPort > 0 {
		phpConfContent += fmt.Sprintf("<FilesMatch \\.php$>\n    SetHandler \"proxy:fcgi://127.0.0.1:%d\"\n</FilesMatch>\n", phpPort)
	}
	phpConfContent += "# End Ostenia PHP\n"
	os.MkdirAll(filepath.Dir(phpConfPath), 0755)
	os.WriteFile(phpConfPath, []byte(phpConfContent), 0644)

	// SSL Config
	if httpsEnabled {
		baseDir := config.GetBaseDir()
		sslDir := filepath.Join(baseDir, "ssl")

		// Ensure certificate exists (Bridge to OpenSSL feature)
		ssl.GenerateRootCA(sslDir) // Ensure CA exists
		ssl.SignCertificate(sslDir, "localhost", sslDir)

		crtPath := strings.ReplaceAll(filepath.Join(sslDir, "localhost.crt"), "\\", "/")
		keyPath := strings.ReplaceAll(filepath.Join(sslDir, "localhost.key"), "\\", "/")

		sslConfContent := fmt.Sprintf(`
# Ostenia SSL Configuration
Listen 443
<VirtualHost _default_:443>
    DocumentRoot "%s"
    ServerName localhost:443
    SSLEngine on
    SSLCertificateFile "%s"
    SSLCertificateKeyFile "%s"
    <Directory "%s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>
# End Ostenia SSL
`, normalizedWWWRoot, crtPath, keyPath, normalizedWWWRoot)

		os.WriteFile(sslConfPath, []byte(sslConfContent), 0644)
	} else {
		os.Remove(sslConfPath)
	}

	return os.WriteFile(confPath, []byte(finalContent), 0644)
}
