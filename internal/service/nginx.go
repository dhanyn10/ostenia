package service

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/nginx.conf.tmpl
var nginxTemplates embed.FS

// NginxConfigData holds the variables for the nginx.conf template.
type NginxConfigData struct {
	Port      int
	WWWRoot   string
	NginxPath string
	PHPPort   int
}

func UpdateNginxConfig(nginxPath string, wwwRoot string, phpPort int, port int) error {
	// Fallback to default port if invalid
	if port <= 0 {
		port = 80
	}

	// Ensure essential directories exist for Windows stability
	os.MkdirAll(filepath.Join(nginxPath, "logs"), 0755)
	os.MkdirAll(filepath.Join(nginxPath, "temp", "client_body_temp"), 0755)
	os.MkdirAll(filepath.Join(nginxPath, "temp", "proxy_temp"), 0755)
	os.MkdirAll(filepath.Join(nginxPath, "temp", "fastcgi_temp"), 0755)

	// Prepare data for template
	data := NginxConfigData{
		Port:      port,
		WWWRoot:   strings.ReplaceAll(wwwRoot, "\\", "/"),
		NginxPath: strings.ReplaceAll(nginxPath, "\\", "/"),
		PHPPort:   phpPort,
	}

	// Load template from embedded filesystem
	tmplBytes, err := nginxTemplates.ReadFile("templates/nginx.conf.tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New("nginx_conf").Parse(string(tmplBytes))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return err
	}

	confPath := filepath.Join(nginxPath, "conf", "nginx.conf")
	return os.WriteFile(confPath, buf.Bytes(), 0644)
}
