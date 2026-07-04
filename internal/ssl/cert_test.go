package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type mockExecutor struct {
	utils.CommandExecutor
	shouldFail bool
}

func (m *mockExecutor) Command(name string, args ...string) *exec.Cmd {
	if m.shouldFail {
		// Return a command that always fails regardless of env override
		// Use a subtest process with a guaranteed exit code 1
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessFail", "--")
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcessOK", "--")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestHelperProcessOK(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestHelperProcessFail(t *testing.T) {
	// This function is invoked as a subprocess – always exit with code 1
	// so that our mock simulates a certutil failure. We do NOT check
	// GO_WANT_HELPER_PROCESS because SafeEnv() in cert.go strips env vars.
	if len(os.Args) > 2 && os.Args[2] == "--" {
		os.Exit(1)
	}
}

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

func TestGetRemainingDaysExpired(t *testing.T) {
	tempDir := t.TempDir()
	caCertPath := filepath.Join(tempDir, "expired.crt")

	origKeySize := RSAKeySize
	RSAKeySize = 1024
	defer func() { RSAKeySize = origKeySize }()

	priv, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-2 * time.Hour),
		NotAfter:     time.Now().Add(-1 * time.Hour), // Expired!
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}

	f, _ := os.Create(caCertPath)
	pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	f.Close()

	days, err := getRemainingDays(caCertPath)
	if err != nil {
		t.Error(err)
	}
	if days != 0 {
		t.Errorf("Expected 0 remaining days for expired certificate, got %d", days)
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

func TestGenerateRootCAWriteError(t *testing.T) {
	invalidDir := filepath.Join("non-existent-directory-xyz-123", "sub")
	err := generateRootCA(invalidDir)
	if err == nil {
		t.Error("Expected error because directory does not exist, got nil")
	}
}

func TestGenerateRootCAKeyWriteError(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "ca.key")
	if err := os.Mkdir(keyPath, 0755); err != nil {
		t.Fatal(err)
	}

	err := generateRootCA(tempDir)
	if err == nil {
		t.Error("Expected error because ca.key is a directory, got nil")
	}
}

func TestSignCertificateWriteError(t *testing.T) {
	tempDir := t.TempDir()

	origKeySize := RSAKeySize
	RSAKeySize = 1024
	defer func() { RSAKeySize = origKeySize }()

	origTrust := TrustRootCAOverride
	TrustRootCAOverride = func(caPath string) error { return nil }
	defer func() { TrustRootCAOverride = origTrust }()

	if err := generateRootCA(tempDir); err != nil {
		t.Fatalf("generateRootCA failed: %v", err)
	}

	invalidDest := filepath.Join("non-existent-directory-xyz-123", "sub")
	err := signCertificate(tempDir, "test.local", invalidDest)
	if err == nil {
		t.Error("Expected error because destination directory does not exist, got nil")
	}
}

func TestSignCertificateKeyWriteError(t *testing.T) {
	tempDir := t.TempDir()

	origKeySize := RSAKeySize
	RSAKeySize = 1024
	defer func() { RSAKeySize = origKeySize }()

	origTrust := TrustRootCAOverride
	TrustRootCAOverride = func(caPath string) error { return nil }
	defer func() { TrustRootCAOverride = origTrust }()

	if err := generateRootCA(tempDir); err != nil {
		t.Fatalf("generateRootCA failed: %v", err)
	}

	domain := "test.local"
	keyPath := filepath.Join(tempDir, domain+".key")
	if err := os.Mkdir(keyPath, 0755); err != nil {
		t.Fatal(err)
	}

	err := signCertificate(tempDir, domain, tempDir)
	if err == nil {
		t.Error("Expected error because domain.key is a directory, got nil")
	}
}

func TestTrustRootCA(t *testing.T) {
	origTrust := TrustRootCAOverride
	TrustRootCAOverride = nil
	defer func() { TrustRootCAOverride = origTrust }()

	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	if runtime.GOOS != "windows" {
		// On non-Windows, TrustRootCA always returns nil (no system trust store interaction)
		err := TrustRootCA("dummy.crt")
		if err != nil {
			t.Errorf("TrustRootCA expected to succeed on non-Windows, got: %v", err)
		}
		return
	}

	// On Windows: test success path (mock succeeds)
	utils.Executor = &mockExecutor{shouldFail: false}
	err := TrustRootCA("dummy.crt")
	if err != nil {
		t.Errorf("TrustRootCA expected to succeed, got error: %v", err)
	}

	// On Windows: test failure path (mock cmd2 fails)
	utils.Executor = &mockExecutor{shouldFail: true}
	err = TrustRootCA("dummy.crt")
	if err == nil {
		t.Error("TrustRootCA expected to fail when certutil returns error, got nil")
	}
}

func TestPublicFunctions(t *testing.T) {
	tempDir := t.TempDir()

	origGen := GenerateRootCAFunc
	GenerateRootCAFunc = func(dir string) error { return nil }
	defer func() { GenerateRootCAFunc = origGen }()

	origDays := GetRemainingDaysFunc
	GetRemainingDaysFunc = func(path string) (int, error) { return 50, nil }
	defer func() { GetRemainingDaysFunc = origDays }()

	origSign := SignCertificateFunc
	SignCertificateFunc = func(caDir, domain, destDir string) error { return nil }
	defer func() { SignCertificateFunc = origSign }()

	if err := GenerateRootCA(tempDir); err != nil {
		t.Error(err)
	}
	if days, err := GetRemainingDays("dummy"); err != nil || days != 50 {
		t.Errorf("days: %d, err: %v", days, err)
	}
	if err := SignCertificate(tempDir, "dom", tempDir); err != nil {
		t.Error(err)
	}
}

