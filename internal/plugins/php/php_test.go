package php

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"testing"
)

type mockHTTPClient struct {
	utils.HTTPClient
	content string
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(m.content)),
	}, nil
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(m.content)),
	}, nil
}

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
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	arch := utils.GetSystemArch()
	mockHTML := `
		<a href="php-8.2.1-Win32-vs16-` + arch + `.zip">php-8.2.1-Win32-vs16-` + arch + `.zip</a>
		<a href="php-8.2.2-Win32-vs16-` + arch + `.zip">php-8.2.2-Win32-vs16-` + arch + `.zip</a>
		<a href="php-8.4.1-Win32-vs16-` + arch + `.zip">php-8.4.1-Win32-vs16-` + arch + `.zip</a>
	`
	utils.Client = &mockHTTPClient{content: mockHTML}

	versions, urlMap := DetectVersions()
	if len(versions) < 2 {
		t.Errorf("Expected at least 2 versions, got %v", versions)
	}

	// Should pick latest patch for each minor
	found84 := false
	found822 := false
	found821 := false
	for _, v := range versions {
		if v == "8.4.1" {
			found84 = true
		}
		if v == "8.2.2" {
			found822 = true
		}
		if v == "8.2.1" {
			found821 = true
		}
	}
	if !found84 || !found822 || found821 {
		t.Errorf("Unexpected versions: found84=%v, found822=%v, found821=%v", found84, found822, found821)
	}

	if _, ok := urlMap["8.2.2"]; !ok {
		t.Error("Expected URL for 8.2.2")
	}
}

func TestGetIcon(t *testing.T) {
	if GetIcon() == "" {
		t.Error("Expected icon")
	}
}

func TestModules(t *testing.T) {
	mods := GetModules()
	if len(mods) == 0 {
		t.Error("Expected modules")
	}

	tempDir := t.TempDir()
	phpPath := tempDir

	// Test GetModuleVersion when not installed
	if GetModuleVersion("Composer", phpPath) != "" {
		t.Error("Expected empty version for missing Composer")
	}

	// Test UninstallModule
	err := UninstallModule("Composer", phpPath)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_ = UninstallModule("Unknown", phpPath)
}

func TestGetModuleVersion(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	tmpDir := t.TempDir()
	composerPhar := filepath.Join(tmpDir, "composer.phar")
	os.WriteFile(composerPhar, []byte(""), 0755)
	phpExe := filepath.Join(tmpDir, "php.exe")
	os.WriteFile(phpExe, []byte(""), 0755)

	t.Run("Success", func(t *testing.T) {
		utils.Executor = &mockExecutor{output: "Composer version 2.5.5"}
		v := GetModuleVersion("Composer", tmpDir)
		if v != "2.5.5" {
			t.Errorf("Expected 2.5.5, got %s", v)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		utils.Executor = &mockExecutor{err: fmt.Errorf("error")}
		v := GetModuleVersion("Composer", tmpDir)
		if v != "" {
			t.Errorf("Expected empty version on error, got %s", v)
		}
	})

	t.Run("UnknownModule", func(t *testing.T) {
		v := GetModuleVersion("Unknown", tmpDir)
		if v != "" {
			t.Errorf("Expected empty version for unknown module, got %s", v)
		}
	})
}

func TestInstallModule(t *testing.T) {
	origClient := utils.Client
	defer func() { utils.Client = origClient }()

	utils.Client = &mockHTTPClient{content: "mock composer phar content"}

	tmpDir := t.TempDir()
	err := InstallModule(nil, nil, "Composer", tmpDir, func(s string, f float64, s2 string) {})
	if err != nil {
		t.Errorf("InstallModule failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "composer.phar")); err != nil {
		t.Error("composer.phar not found after install")
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "composer.bat")); err != nil {
		t.Error("composer.bat not found after install")
	}

	// Test unknown module
	err = InstallModule(nil, nil, "Unknown", tmpDir, nil)
	if err == nil {
		t.Error("Expected error for unknown module")
	}
}
