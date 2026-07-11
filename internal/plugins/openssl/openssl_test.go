package openssl

import (
	"fmt"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"testing"
)

type mockExecutor struct {
	output string
	err    error
}

func (m *mockExecutor) Command(name string, arg ...string) *exec.Cmd {
	argList := []string{"-test.run=TestHelperProcess", "--", name}
	argList = append(argList, arg...)
	cmd := exec.Command(os.Args[0], argList...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "MOCK_OUTPUT="+m.output)
	if m.err != nil {
		cmd.Env = append(cmd.Env, "MOCK_EXIT_CODE=1")
	}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("MOCK_OUTPUT") != "" {
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_OUTPUT"))
	}
	if os.Getenv("MOCK_EXIT_CODE") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
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

func TestDetectInstalledVersion(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	utils.Executor = &mockExecutor{output: "OpenSSL 3.0.7 1 Nov 2022"}

	version := DetectInstalledVersion()
	if version != "3.0.7" {
		t.Errorf("Expected 3.0.7, got %s", version)
	}
}

func TestFindExecutables(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	// Mock 'where openssl' output
	utils.Executor = &mockExecutor{output: "C:\\bin\\openssl.exe\n"}

	// We can't easily test the full findExecutables because it checks os.Stat
	// but we can at least run it.
	_ = findExecutables()
}

func TestVersionFromExecutable(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	t.Run("Success", func(t *testing.T) {
		utils.Executor = &mockExecutor{output: "OpenSSL 1.1.1q  17 Aug 2022"}
		v := versionFromExecutable("openssl")
		if v != "1.1.1q" {
			t.Errorf("Expected 1.1.1q, got %s", v)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		utils.Executor = &mockExecutor{err: fmt.Errorf("error")}
		v := versionFromExecutable("invalid")
		if v != "" {
			t.Errorf("Expected empty version on error, got %s", v)
		}
	})

	t.Run("ShortOutput", func(t *testing.T) {
		utils.Executor = &mockExecutor{output: "OpenSSL"}
		v := versionFromExecutable("openssl")
		if v != "" {
			t.Errorf("Expected empty version on short output, got %s", v)
		}
	})
}
