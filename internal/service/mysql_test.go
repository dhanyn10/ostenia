package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMySQLConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mysql_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mysqlBaseDir := filepath.Join(tmpDir, "mysql")
	dataDir := filepath.Join(mysqlBaseDir, "data")
	tmpMySQLDir := filepath.Join(mysqlBaseDir, "tmp")

	err = UpdateMySQLConfig(mysqlBaseDir, dataDir, tmpMySQLDir, 3306)
	if err != nil {
		t.Fatalf("UpdateMySQLConfig failed: %v", err)
	}

	iniPath := filepath.Join(mysqlBaseDir, "my.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		t.Fatal("my.ini was not created")
	}

	content, _ := os.ReadFile(iniPath)
	if !strings.Contains(string(content), "port=3306") {
		t.Error("Expected port=3306 in my.ini")
	}
}
