package ssl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSSL_Complete(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("GenerateRootCA", func(t *testing.T) {
		err := GenerateRootCA(tmpDir)
		if err != nil {
			t.Fatalf("GenerateRootCA failed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(tmpDir, "ca.crt")); err != nil {
			t.Error("ca.crt not created")
		}
	})

	t.Run("GetRemainingDays", func(t *testing.T) {
		days, err := GetRemainingDays(filepath.Join(tmpDir, "ca.crt"))
		if err != nil {
			t.Errorf("GetRemainingDays failed: %v", err)
		}
		if days < 0 {
			t.Errorf("Unexpected days: %d", days)
		}
	})

	t.Run("SignCertificate", func(t *testing.T) {
		err := SignCertificate(tmpDir, "test.local", tmpDir)
		if err != nil {
			t.Errorf("SignCertificate failed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(tmpDir, "test.local.crt")); err != nil {
			t.Error("test.local.crt not created")
		}
	})

	t.Run("TrustRootCA", func(t *testing.T) {
		// Just call it to cover non-windows path
		_ = TrustRootCA(filepath.Join(tmpDir, "ca.crt"))
	})
}
