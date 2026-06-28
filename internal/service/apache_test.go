package service

import (
	"strings"
	"testing"
)

func TestApacheVHost(t *testing.T) {
	vhost := GenerateVHost("myproj", "C:/www/myproj", 8080)
	if !strings.Contains(vhost, "ServerName myproj.test") {
		t.Error("Expected ServerName myproj.test")
	}
	if !strings.Contains(vhost, "DocumentRoot \"C:/www/myproj\"") {
		t.Error("Expected correct DocumentRoot")
	}
}

func TestApacheProxyVHost(t *testing.T) {
	vhost := GenerateProxyVHost("proxy", 3000, 80, false, "")
	if !strings.Contains(vhost, "ProxyPass / http://127.0.0.1:3000/") {
		t.Error("Expected correct ProxyPass")
	}
}
