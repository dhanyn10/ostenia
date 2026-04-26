package service

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/ssl"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/nginx.conf.tmpl
var nginxTemplates embed.FS

type NginxConfigData struct {
	Port         int
	WWWRoot      string
	NginxPath    string
	PHPPort      int
	HTTPSEnabled bool
	CertFile     string
	KeyFile      string
}

func UpdateNginxConfig(nginxPath string, wwwRoot string, phpPort int, port int, httpsEnabled bool) error {
	if port <= 0 { port = 80 }

	// Essential directories
	os.MkdirAll(filepath.Join(nginxPath, "logs"), 0755)
	os.MkdirAll(filepath.Join(nginxPath, "temp", "client_body_temp"), 0755)
	os.MkdirAll(filepath.Join(nginxPath, "temp", "proxy_temp"), 0755)
	os.MkdirAll(filepath.Join(nginxPath, "temp", "fastcgi_temp"), 0755)

	baseDir := config.GetBaseDir()
	sslDir := filepath.Join(baseDir, "ssl")

	certFile := ""
	keyFile := ""

	if httpsEnabled {
		// Ensure SSL dir exists
		os.MkdirAll(sslDir, 0755)

		// Sign certificate (Bridge to OpenSSL feature)
		// We ignore error here but will check file existence
		_ = ssl.SignCertificate(sslDir, "localhost", sslDir)

		certPath := filepath.Join(sslDir, "localhost.crt")
		keyPath := filepath.Join(sslDir, "localhost.key")

		if _, err := os.Stat(certPath); err == nil {
			certFile = strings.ReplaceAll(certPath, "\\", "/")
			keyFile = strings.ReplaceAll(keyPath, "\\", "/")
		} else {
			// Fallback: If cert failed to generate, disable HTTPS to prevent Nginx crash
			httpsEnabled = false
			fmt.Printf("[Nginx] SSL Cert not found, disabling HTTPS to prevent crash\n")
		}
	}

	data := NginxConfigData{
		Port:         port,
		WWWRoot:      strings.ReplaceAll(wwwRoot, "\\", "/"),
		NginxPath:    strings.ReplaceAll(nginxPath, "\\", "/"),
		PHPPort:      phpPort,
		HTTPSEnabled: httpsEnabled,
		CertFile:     certFile,
		KeyFile:      keyFile,
	}

	tmplBytes, err := nginxTemplates.ReadFile("templates/nginx.conf.tmpl")
	if err != nil { return err }

	tmpl, err := template.New("nginx_conf").Parse(string(tmplBytes))
	if err != nil { return err }

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil { return err }

	confPath := filepath.Join(nginxPath, "conf", "nginx.conf")
	return os.WriteFile(confPath, buf.Bytes(), 0644)
}
