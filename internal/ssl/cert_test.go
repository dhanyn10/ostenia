package ssl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCertLogic(t *testing.T) {
	// This test verifies the ORCHESTRATION of cert generation without heavy RSA
	// by mocking the actual generateRootCA implementation.

	tempDir := t.TempDir()

	origGen := GenerateRootCAFunc
	defer func() { GenerateRootCAFunc = origGen }()

	called := false
	GenerateRootCAFunc = func(dir string) error {
		called = true
		// Mock file creation
		os.WriteFile(filepath.Join(dir, "ca.crt"), []byte("mock cert"), 0644)
		return nil
	}

	err := GenerateRootCA(tempDir)
	if err != nil {
		t.Fatalf("GenerateRootCA failed: %v", err)
	}

	if !called {
		t.Error("GenerateRootCAFunc was not called")
	}

	if _, err := os.Stat(filepath.Join(tempDir, "ca.crt")); err != nil {
		t.Error("Mock ca.crt not found")
	}
}

func TestRealCertGenerationShort(t *testing.T) {
	// Keep one minimal "real" test but with small key size for coverage of the actual crypto code
	origRSA := RSAKeySize
	RSAKeySize = 1024 // Small for speed, only for testing. 1024 is the minimum for some versions.
	defer func() { RSAKeySize = origRSA }()

	TrustRootCAOverride = func(caPath string) error { return nil }
	defer func() { TrustRootCAOverride = nil }()

	tempDir := t.TempDir()

	t.Run("RealGeneration", func(t *testing.T) {
		err := generateRootCA(tempDir) // Call the internal implementation
		if err != nil {
			t.Fatalf("generateRootCA failed: %v", err)
		}

		days, err := getRemainingDays(filepath.Join(tempDir, "ca.crt"))
		if err != nil {
			t.Fatalf("getRemainingDays failed: %v", err)
		}
		if days < 360 {
			t.Errorf("Expected ~365 days, got %d", days)
		}
	})
}
