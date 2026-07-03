package ssl

import (
	"os"
	"path/filepath"
	"testing"
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
