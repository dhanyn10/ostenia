package ssl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCertReal(t *testing.T) {
	tempDir := t.TempDir()

	// Speed up tests by using smaller RSA key size
	origKeySize := RSAKeySize
	RSAKeySize = 1024
	defer func() { RSAKeySize = origKeySize }()

	// Mock TrustRootCA to do nothing
	origTrust := TrustRootCAOverride
	TrustRootCAOverride = func(caPath string) error { return nil }
	defer func() { TrustRootCAOverride = origTrust }()

	// Test generateRootCA
	err := generateRootCA(tempDir)
	if err != nil {
		t.Fatalf("generateRootCA failed: %v", err)
	}

	caCertPath := filepath.Join(tempDir, "ca.crt")
	caKeyPath := filepath.Join(tempDir, "ca.key")

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		t.Errorf("ca.crt not created")
	}
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		t.Errorf("ca.key not created")
	}

	// Test generateRootCA again (should return nil early)
	err = generateRootCA(tempDir)
	if err != nil {
		t.Errorf("generateRootCA second call failed: %v", err)
	}

	// Test signCertificate
	domain := "test.local"
	err = signCertificate(tempDir, domain, tempDir)
	if err != nil {
		t.Fatalf("signCertificate failed: %v", err)
	}

	certPath := filepath.Join(tempDir, domain+".crt")
	keyPath := filepath.Join(tempDir, domain+".key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("%s not created", certPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("%s not created", keyPath)
	}

	// Test signCertificate again (should return nil early)
	err = signCertificate(tempDir, domain, tempDir)
	if err != nil {
		t.Errorf("signCertificate second call failed: %v", err)
	}

	// Test getRemainingDays
	days, err := getRemainingDays(caCertPath)
	if err != nil {
		t.Errorf("getRemainingDays failed: %v", err)
	}
	if days <= 0 {
		t.Errorf("Expected positive remaining days, got %d", days)
	}
}

func TestGetRemainingDaysError(t *testing.T) {
	// File not found
	_, err := getRemainingDays("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Invalid PEM
	tempDir := t.TempDir()
	invalidPath := filepath.Join(tempDir, "invalid.crt")
	os.WriteFile(invalidPath, []byte("not a pem"), 0644)
	_, err = getRemainingDays(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}

	// Invalid certificate data
	invalidCertPath := filepath.Join(tempDir, "invalid_cert.crt")
	os.WriteFile(invalidCertPath, []byte("-----BEGIN CERTIFICATE-----\ninvalid\n-----END CERTIFICATE-----"), 0644)
	_, err = getRemainingDays(invalidCertPath)
	if err == nil {
		t.Error("Expected error for invalid certificate data")
	}
}

func TestSignCertificateError(t *testing.T) {
	tempDir := t.TempDir()

	// Missing CA cert
	err := signCertificate(tempDir, "test.local", tempDir)
	if err == nil {
		t.Error("Expected error for missing CA cert")
	}

	// Create dummy CA cert but missing CA key
	os.WriteFile(filepath.Join(tempDir, "ca.crt"), []byte("dummy"), 0644)
	err = signCertificate(tempDir, "test.local", tempDir)
	if err == nil {
		t.Error("Expected error for missing CA key")
	}

	// Invalid CA cert
	os.WriteFile(filepath.Join(tempDir, "ca.key"), []byte("dummy"), 0644)
	err = signCertificate(tempDir, "test.local", tempDir)
	if err == nil {
		t.Error("Expected error for invalid CA cert")
	}
}

func TestTrustRootCA(t *testing.T) {
	// TrustRootCA is mostly windows specific and already tested via TrustRootCAOverride in TestCertReal
	// But let's call it without override on non-windows to ensure it doesn't crash
	origTrust := TrustRootCAOverride
	TrustRootCAOverride = nil
	defer func() { TrustRootCAOverride = origTrust }()

	err := TrustRootCA("dummy.crt")
	if err != nil {
		t.Errorf("TrustRootCA failed on current OS: %v", err)
	}
}

func TestPublicFunctions(t *testing.T) {
	tempDir := t.TempDir()

	// Mock them
	origGen := GenerateRootCAFunc
	GenerateRootCAFunc = func(dir string) error { return nil }
	defer func() { GenerateRootCAFunc = origGen }()

	origDays := GetRemainingDaysFunc
	GetRemainingDaysFunc = func(path string) (int, error) { return 50, nil }
	defer func() { GetRemainingDaysFunc = origDays }()

	origSign := SignCertificateFunc
	SignCertificateFunc = func(caDir, domain, destDir string) error { return nil }
	defer func() { SignCertificateFunc = origSign }()

	if err := GenerateRootCA(tempDir); err != nil { t.Error(err) }
	if days, err := GetRemainingDays("dummy"); err != nil || days != 50 { t.Errorf("days: %d, err: %v", days, err) }
	if err := SignCertificate(tempDir, "dom", tempDir); err != nil { t.Error(err) }
}
