package ssl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSSLGeneration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ssl_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test GenerateRootCA (Note: TrustRootCA will likely fail or do nothing on Linux, which is fine)
	err = GenerateRootCA(tmpDir)
	if err != nil {
		t.Errorf("GenerateRootCA failed: %v", err)
	}

	caCertPath := filepath.Join(tmpDir, "ca.crt")
	caKeyPath := filepath.Join(tmpDir, "ca.key")

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		t.Errorf("ca.crt not created")
	}
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		t.Errorf("ca.key not created")
	}

	// Test GetRemainingDays
	days, err := GetRemainingDays(caCertPath)
	if err != nil {
		t.Errorf("GetRemainingDays failed: %v", err)
	}
	if days <= 0 {
		t.Errorf("Expected positive remaining days, got %d", days)
	}

	// Test SignCertificate
	domain := "example.test"
	err = SignCertificate(tmpDir, domain, tmpDir)
	if err != nil {
		t.Errorf("SignCertificate failed: %v", err)
	}

	domainCertPath := filepath.Join(tmpDir, domain+".crt")
	domainKeyPath := filepath.Join(tmpDir, domain+".key")

	if _, err := os.Stat(domainCertPath); os.IsNotExist(err) {
		t.Errorf("%s.crt not created", domain)
	}
	if _, err := os.Stat(domainKeyPath); os.IsNotExist(err) {
		t.Errorf("%s.key not created", domain)
	}
}

func TestGetRemainingDaysError(t *testing.T) {
	_, err := GetRemainingDays("non_existent.crt")
	if err == nil {
		t.Error("Expected error for non-existent cert, got nil")
	}

	// Test with invalid PEM
	tmpFile, _ := os.CreateTemp("", "invalid_cert")
	defer os.Remove(tmpFile.Name())
	os.WriteFile(tmpFile.Name(), []byte("not a cert"), 0644)

	_, err = GetRemainingDays(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid PEM, got nil")
	}
}
