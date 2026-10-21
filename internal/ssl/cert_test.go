package ssl

import (
	"crypto/rand"
	"crypto/x509"
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"testing"
	"time"
)

func TestCertMocked(t *testing.T) {
	tempDir := t.TempDir()

	// Mock GenerateRootCA
	origGen := GenerateRootCAFunc
	GenerateRootCAFunc = func(dir string) error {
		os.WriteFile(filepath.Join(dir, "ca.crt"), []byte("mock cert"), 0644)
		return nil
	}
	defer func() { GenerateRootCAFunc = origGen }()

	err := GenerateRootCA(tempDir)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Mock SignCertificate
	origSign := SignCertificateFunc
	SignCertificateFunc = func(caDir, domain, destDir string) error {
		os.WriteFile(filepath.Join(destDir, domain+".crt"), []byte("mock signed cert"), 0644)
		return nil
	}
	defer func() { SignCertificateFunc = origSign }()

	err = SignCertificate(tempDir, "test.local", tempDir)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	// Mock GetRemainingDays
	origDays := GetRemainingDaysFunc
	GetRemainingDaysFunc = func(path string) (int, error) {
		return 100, nil
	}
	defer func() { GetRemainingDaysFunc = origDays }()

	days, err := GetRemainingDays("dummy")
	if err != nil || days != 100 {
		t.Errorf("Expected 100 days, got %d (err: %v)", days, err)
	}
}

func TestRealCertGeneration(t *testing.T) {
	tempDir := t.TempDir()

	// Speed up tests by using smaller key size
	origKeySize := RSAKeySize
	RSAKeySize = 1024
	defer func() { RSAKeySize = origKeySize }()

	// Ensure TrustRootCA is mocked
	origTrust := TrustRootCAOverride
	TrustRootCAOverride = func(path string) error { return nil }
	defer func() { TrustRootCAOverride = origTrust }()

	// 1. Generate Root CA
	err := generateRootCA(tempDir)
	if err != nil {
		t.Fatalf("Failed to generate Root CA: %v", err)
	}

	// Test re-generation (should skip)
	err = generateRootCA(tempDir)
	if err != nil {
		t.Errorf("generateRootCA failed on second call: %v", err)
	}

	caCertPath := filepath.Join(tempDir, "ca.crt")
	caKeyPath := filepath.Join(tempDir, "ca.key")

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		t.Error("ca.crt was not created")
	}
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		t.Error("ca.key was not created")
	}

	// 2. Test GetRemainingDays
	days, err := getRemainingDays(caCertPath)
	if err != nil {
		t.Errorf("getRemainingDays failed: %v", err)
	}
	if days < 364 || days > 366 {
		t.Errorf("Expected ~365 days remaining, got %d", days)
	}

	// 3. Sign a certificate
	domain := "test.example.com"
	err = signCertificate(tempDir, domain, tempDir)
	if err != nil {
		t.Fatalf("Failed to sign certificate: %v", err)
	}

	// Test re-signing (should skip)
	err = signCertificate(tempDir, domain, tempDir)
	if err != nil {
		t.Errorf("signCertificate failed on second call: %v", err)
	}

	certPath := filepath.Join(tempDir, domain+".crt")
	keyPath := filepath.Join(tempDir, domain+".key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Error("domain.crt was not created")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("domain.key was not created")
	}

	// 4. Verify the signed certificate
	cert, err := LoadCertificate(certPath)
	if err != nil {
		t.Fatalf("Failed to load signed certificate: %v", err)
	}

	if cert.Subject.CommonName != domain {
		t.Errorf("Expected CN %s, got %s", domain, cert.Subject.CommonName)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test GeneratePrivateKey
	priv, err := GeneratePrivateKey(1024)
	if err != nil {
		t.Errorf("GeneratePrivateKey failed: %v", err)
	}
	if priv.PublicKey.N.BitLen() < 1024 {
		t.Errorf("Expected 1024 bit key, got %d", priv.PublicKey.N.BitLen())
	}

	// Test Templates
	rootTemplate := CreateRootCATemplate()
	if !rootTemplate.IsCA {
		t.Error("Root template should have IsCA=true")
	}

	caCert := &x509.Certificate{SubjectKeyId: []byte("mock-skid")}
	eeTemplate := CreateEndEntityTemplate("example.com", caCert)
	if eeTemplate.IsCA {
		t.Error("End entity template should have IsCA=false")
	}
	if string(eeTemplate.AuthorityKeyId) != string(caCert.SubjectKeyId) {
		t.Error("AuthorityKeyId mismatch in end entity template")
	}

	// Test PEM Encoding
	keyPEM := EncodeKeyToPEM(priv)
	if len(keyPEM) == 0 {
		t.Error("EncodeKeyToPEM returned empty data")
	}
}

func TestLoadErrors(t *testing.T) {
	tempDir := t.TempDir()
	nonExistent := filepath.Join(tempDir, "missing")

	_, err := LoadCertificate(nonExistent)
	if err == nil {
		t.Error("Expected error loading non-existent certificate")
	}

	_, err = LoadPrivateKey(nonExistent)
	if err == nil {
		t.Error("Expected error loading non-existent private key")
	}

	invalidPath := filepath.Join(tempDir, "invalid.txt")
	os.WriteFile(invalidPath, []byte("not a pem"), 0644)

	_, err = LoadCertificate(invalidPath)
	if err == nil {
		t.Error("Expected error loading invalid PEM as certificate")
	}

	_, err = LoadPrivateKey(invalidPath)
	if err == nil {
		t.Error("Expected error loading invalid PEM as private key")
	}
}

func TestGenerateAndSignErrors(t *testing.T) {
	// Test generateRootCA with an unwritable/invalid path
	err := generateRootCA("/invalid-path-foo-bar/baz")
	if err == nil {
		t.Error("Expected error for unwritable root CA destination path")
	}

	// Test signCertificate with non-existent CA dir
	err = signCertificate("/invalid-ca-dir", "dummy.com", t.TempDir())
	if err == nil {
		t.Error("Expected error for non-existent CA directory in signCertificate")
	}

	// Test signCertificate with valid CA dir but missing ca.key
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "ca.crt"), []byte("dummy-ca-cert"), 0644)
	err = signCertificate(tempDir, "dummy.com", tempDir)
	if err == nil {
		t.Error("Expected error for missing ca.key in signCertificate")
	}
}

func TestTrustRootCA(t *testing.T) {
	// Test override (cross-platform)
	called := false
	origTrust := TrustRootCAOverride
	TrustRootCAOverride = func(path string) error {
		called = true
		return nil
	}
	defer func() { TrustRootCAOverride = origTrust }()

	TrustRootCA("some/path")
	if !called {
		t.Error("TrustRootCAOverride was not called")
	}

	// Reset override to test real logic
	TrustRootCAOverride = nil

	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &testutil.MockExecutor{Output: ""}

	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()

	// Test Windows execution branch with MockExecutor
	RuntimeGOOS = "windows"
	err := TrustRootCA("any/path")
	if err != nil {
		t.Errorf("TrustRootCA failed on windows branch: %v", err)
	}

	// Test non-Windows branch
	RuntimeGOOS = "linux"
	err = TrustRootCA("any/path")
	if err != nil {
		t.Errorf("TrustRootCA failed on non-windows branch: %v", err)
	}
}

func TestSignCertificate_Errors(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Missing ca.crt
	err := signCertificate(tempDir, "test.local", tempDir)
	if err == nil {
		t.Error("Expected error when ca.crt is missing")
	}

	// 2. ca.crt exists but ca.key is missing
	_ = os.WriteFile(filepath.Join(tempDir, "ca.crt"), []byte("invalid cert"), 0600)
	err = signCertificate(tempDir, "test.local", tempDir)
	if err == nil {
		t.Error("Expected error when ca.key is missing")
	}
}

func TestGetRemainingDaysExpired(t *testing.T) {
	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "expired.crt")

	priv, _ := GeneratePrivateKey(1024)
	template := CreateRootCATemplate()
	template.NotAfter = time.Now().Add(-24 * time.Hour)

	der, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	os.WriteFile(certPath, EncodeCertToPEM(der), 0644)

	days, err := getRemainingDays(certPath)
	if err != nil {
		t.Errorf("getRemainingDays failed: %v", err)
	}
	if days != 0 {
		t.Errorf("Expected 0 days for expired cert, got %d", days)
	}
}
