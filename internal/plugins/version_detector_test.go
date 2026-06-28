package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetInstalledVersionPaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugins_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	category := "php"
	binDir := filepath.Join(tmpDir, "bin", category)
	versionDir := filepath.Join(binDir, "8.2.0")
	os.MkdirAll(versionDir, 0755)

	checkFile := "php-cgi.exe"
	os.WriteFile(filepath.Join(versionDir, checkFile), []byte("dummy"), 0644)

	versions := GetInstalledVersionPaths(tmpDir, category, checkFile)
	if len(versions) == 0 {
		t.Error("Expected 1 version, got 0")
	}
	if _, ok := versions["8.2.0"]; !ok {
		t.Errorf("Expected version 8.2.0 to be found, got %v", versions)
	}

	// Test Apache special case
	categoryApache := "apache"
	apacheBinDir := filepath.Join(tmpDir, "bin", categoryApache)
	apacheVerDir := filepath.Join(apacheBinDir, "2.4.54")
	apacheNestedBinDir := filepath.Join(apacheVerDir, "Apache24", "bin")
	os.MkdirAll(apacheNestedBinDir, 0755)

	apacheCheckFile := "bin/httpd.exe"
	os.WriteFile(filepath.Join(apacheNestedBinDir, "httpd.exe"), []byte("dummy"), 0644)

	versionsApache := GetInstalledVersionPaths(tmpDir, categoryApache, apacheCheckFile)
	if _, ok := versionsApache["2.4.54"]; !ok {
		t.Errorf("Expected version 2.4.54 for Apache, got %v", versionsApache)
	}
}
