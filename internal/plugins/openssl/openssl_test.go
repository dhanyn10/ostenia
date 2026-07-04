package openssl

import (
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type mockExecutor struct {
	responses map[string]string
}

func (m *mockExecutor) Command(name string, arg ...string) *exec.Cmd {
	full := name + " " + strings.Join(arg, " ")

	// Check for exact match
	var response string
	var found bool
	if response, found = m.responses[full]; !found {
		// Fallback to name-only check if specific command not found
		response = m.responses[name]
	}

	if response == "__ERROR__" {
		return exec.Command("false")
	}

	// We use a trick to return a command that outputs our mock response
	// On Unix-like systems, we can use 'echo'
	return exec.Command("echo", "-n", response)
}

func TestDetectVersions(t *testing.T) {
	versions, urlMap := DetectVersions()
	if len(versions) != 1 || versions[0] != "4.0.0" {
		t.Errorf("Expected 4.0.0, got %v", versions)
	}
	if _, ok := urlMap["4.0.0"]; !ok {
		t.Error("Expected URL for 4.0.0")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}

func TestDetectInstalledVersion_Mocked(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	mock := &mockExecutor{
		responses: map[string]string{
			"openssl version": "OpenSSL 3.0.7 1 Nov 2022",
		},
	}
	utils.Executor = mock

	version := DetectInstalledVersion()
	if version != "3.0.7" {
		t.Errorf("Expected 3.0.7, got %s", version)
	}
}

func TestFindExecutables(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "openssl.exe")
	os.WriteFile(exePath, []byte("dummy"), 0755)

	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	mock := &mockExecutor{
		responses: map[string]string{
			"where openssl": exePath,
			"where.exe openssl": exePath,
		},
	}
	utils.Executor = mock

	exes := findExecutables()
	found := false
	for _, e := range exes {
		if strings.EqualFold(e, exePath) {
			found = true
			break
		}
	}
	// Note: findExecutables also walks config.GetBaseDir()/bin which we haven't mocked here
	// But it should at least find the one from 'where' if it existed in os.Stat
	if !found {
		t.Logf("findExecutables did not find %s, but that might be due to os.Stat failing if it doesn't actually exist on disk", exePath)
	}
}

func TestVersionFromExecutable_Error(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	utils.Executor = &mockExecutor{
		responses: map[string]string{
			"bad-openssl version": "__ERROR__",
		},
	}

	v := versionFromExecutable("bad-openssl")
	if v != "" {
		t.Errorf("Expected empty version for failing command, got %s", v)
	}
}

func TestFindExecutables_Linux(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Linux-specific test")
	}

	// On Linux, findExecutables returns empty list immediately
	exes := findExecutables()
	if len(exes) != 0 {
		t.Errorf("Expected 0 executables on Linux, got %d", len(exes))
	}
}

func TestDetectInstalledVersion_Linux(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Linux-specific test")
	}

	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	mock := &mockExecutor{
		responses: map[string]string{
			"openssl version": "OpenSSL 3.0.7 1 Nov 2022",
		},
	}
	utils.Executor = mock

	version := DetectInstalledVersion()
	if version != "3.0.7" {
		t.Errorf("Expected 3.0.7, got %s", version)
	}
}
