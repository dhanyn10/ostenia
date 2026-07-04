package openssl

import (
	"fmt"
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

	var response string
	var found bool
	if response, found = m.responses[full]; !found {
		response = m.responses[name]
	}

	exitCode := "0"
	if response == "__ERROR__" {
		exitCode = "1"
		response = ""
	}

	argList := []string{"-test.run=TestHelperProcess", "--"}
	argList = append(argList, arg...)
	cmd := exec.Command(os.Args[0], argList...)
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"MOCK_OUTPUT="+response,
		"MOCK_EXIT_CODE="+exitCode,
	)
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("MOCK_OUTPUT") != "" {
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_OUTPUT"))
	}
	exitCode := 0
	fmt.Sscanf(os.Getenv("MOCK_EXIT_CODE"), "%d", &exitCode)
	os.Exit(exitCode)
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
	if !found {
		t.Logf("findExecutables did not find %s", exePath)
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
