package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func UpdateNginxConfig(nginxPath string, wwwRoot string, phpPort int, port int) error {
	confPath := filepath.Join(nginxPath, "conf", "nginx.conf")

	normalizedWWWRoot := strings.ReplaceAll(wwwRoot, "\\", "/")

	// Basic optimized Nginx configuration
	content := fmt.Sprintf(`
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile        on;
    keepalive_timeout  65;

    server {
        listen       %d;
        server_name  localhost;
        root         "%s";
        index        index.php index.html index.htm;

        location / {
            try_files $uri $uri/ /index.php?$query_string;
        }

        # PHP-FPM / FastCGI pass
        location ~ \.php$ {
            fastcgi_pass   127.0.0.1:%d;
            fastcgi_index  index.php;
            fastcgi_param  SCRIPT_FILENAME  $document_root$fastcgi_script_name;
            include        fastcgi_params;
        }

        # Fallback to Ostenia default if root is empty (simulated via error_page or rewrite)
        error_page 404 /index.php;
    }
}
`, port, normalizedWWWRoot, phpPort)

	return os.WriteFile(confPath, []byte(content), 0644)
}
