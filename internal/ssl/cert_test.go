package ssl

import (
	"os"
	"path/filepath"
	"testing"
	"ostenia/internal/plugins/utils"
	"os/exec"
)

type mockExecutor struct {
	utils.CommandExecutor
}

func (m *mockExecutor) Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--")
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestCert(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &mockExecutor{}

	origRSA := RSAKeySize
	RSAKeySize = 1024
	defer func() { RSAKeySize = origRSA }()

	tempDir := t.TempDir()

	t.Run("GenerateRootCA", func(t *testing.T) {
		err := GenerateRootCA(tempDir)
		if err != nil {
			t.Fatalf("GenerateRootCA failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(tempDir, "ca.crt")); err != nil {
			t.Error("ca.crt not created")
		}
		if _, err := os.Stat(filepath.Join(tempDir, "ca.key")); err != nil {
			t.Error("ca.key not created")
		}

		// Second call should return nil (already exists)
		err = GenerateRootCA(tempDir)
		if err != nil {
			t.Errorf("GenerateRootCA second call failed: %v", err)
		}
	})

	t.Run("GetRemainingDays", func(t *testing.T) {
		days, err := GetRemainingDays(filepath.Join(tempDir, "ca.crt"))
		if err != nil {
			t.Fatalf("GetRemainingDays failed: %v", err)
		}
		if days < 360 {
			t.Errorf("Expected ~365 days, got %d", days)
		}

		_, err = GetRemainingDays("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("SignCertificate", func(t *testing.T) {
		err := SignCertificate(tempDir, "test.local", tempDir)
		if err != nil {
			t.Fatalf("SignCertificate failed: %v", err)
		}

		if _, err := os.Stat(filepath.Join(tempDir, "test.local.crt")); err != nil {
			t.Error("test.local.crt not created")
		}
		if _, err := os.Stat(filepath.Join(tempDir, "test.local.key")); err != nil {
			t.Error("test.local.key not created")
		}

		// Test error when CA is missing
		err = SignCertificate("non-existent", "fail.local", tempDir)
		if err == nil {
			t.Error("Expected error when CA is missing")
		}
	})
}
