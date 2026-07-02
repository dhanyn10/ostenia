package ssl

import (
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"testing"
)

type sslMockExecutor struct {
	utils.CommandExecutor
}

func (m *sslMockExecutor) Command(name string, args ...string) *exec.Cmd {
	// Return a dummy command that does nothing
	return exec.Command("go", "version")
}

func TestSSL_Complete(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &sslMockExecutor{}

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
