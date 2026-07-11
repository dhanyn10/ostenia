package service

import (
	"context"
	"io"
	"os"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/testutil"
	"path/filepath"
	"testing"
	"time"
)

type mockRuntime struct {
	interfaces.Runtime
}

func (m *mockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {}

func TestHelperProcess(t *testing.T) {
	testutil.HelperProcess(t)
}

func TestOrchestrator_Complete(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orch := NewOrchestrator(ctx)
	orch.SetRuntime(&mockRuntime{})

	t.Run("TabManagement", func(t *testing.T) {
		orch.SetActiveTab("plugins")
		orch.SetActiveTab("activity")
		orch.RequestRefresh()
	})

	t.Run("ServiceStatus", func(t *testing.T) {
		orch.updateCache("Apache", ServiceDetailedInfo{Status: "Running", PID: 123, Port: 80})
		if !orch.IsRunning("Apache") {
			t.Error("Expected Apache to be running")
		}
		info := orch.GetDetailedInfo("Apache")
		if info.PID != 123 {
			t.Errorf("Expected PID 123, got %d", info.PID)
		}

		info = orch.GetDetailedInfo("Unknown")
		if info.Status != "Stopped" {
			t.Errorf("Expected Stopped for unknown service, got %s", info.Status)
		}
	})

	t.Run("UpdateServiceInfoMocks", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		findOsteniaPIDsOverride = func(exeName string) []int {
			if exeName == "httpd.exe" {
				return []int{1234}
			}
			return []int{}
		}
		defer func() { findOsteniaPIDsOverride = nil }()

		orch.updateServiceInfo("Apache")
		info := orch.GetDetailedInfo("Apache")
		if info.Status != "Running" || info.PID != 1234 {
			t.Errorf("Expected Apache to be Running with PID 1234, got %s (PID %d)", info.Status, info.PID)
		}

		orch.updateServiceInfo("Node.js")
		orch.updateServiceInfo("MySQL")
		orch.updateServiceInfo("Nginx")
		orch.updateServiceInfo("PHP")
		orch.updateServiceInfo("Python")
		orch.updateServiceInfo("HeidiSQL")
		orch.updateServiceInfo("OpenSSL")
	})

	t.Run("StartStopService", func(t *testing.T) {
		utils.Executor = &testutil.MockExecutor{Output: ""}

		err := orch.StartService("TestService", os.Args[0], []string{"-test.run=TestHelperProcess", "--"}, "")
		if err != nil {
			t.Errorf("StartService failed: %v", err)
		}

		orch.StopAll()
		// Small sleep to allow goroutines to finish and file locks to be released on Windows
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("WatcherAndRefresh", func(t *testing.T) {
		orch.RequestRefresh()
		orch.performWatcherCheck([]string{"Apache"})
	})

	t.Run("ServiceTrackingAndCleanup", func(t *testing.T) {
		info := &ServiceDetailedInfo{Name: "Apache", Status: "Running", Port: 80}
		orch.ensureServiceTracked("Apache", info, nil, false)
		orch.cleanupUntrackedService("Apache", "/tmp")
	})

	t.Run("NetstatAndPorts", func(t *testing.T) {
		findPortsByPIDExactOverride = func(pid int) []int {
			return []int{80}
		}
		defer func() { findPortsByPIDExactOverride = nil }()

		ports := findPortsByPIDExact(1234)
		if len(ports) == 0 || ports[0] != 80 {
			t.Errorf("Expected port 80, got %v", ports)
		}

		info := &ServiceDetailedInfo{Name: "Apache"}
		orch.updateServicePorts(info, []int{1234}, nil, false)
		if info.Port != 80 {
			t.Errorf("Expected port 80, got %d", info.Port)
		}
	})
}

func TestParseNetstatOutput(t *testing.T) {
	output := "  TCP    0.0.0.0:80             0.0.0.0:0              LISTENING       1234\n"
	ports := parseNetstatOutput(output, 1234)
	if len(ports) != 1 || ports[0] != 80 {
		t.Errorf("Expected port 80, got %v", ports)
	}
}

func TestParseWmicOutput(t *testing.T) {
	output := "Node,ExecutablePath,ProcessId\nMYPC,C:\\bin\\apache\\httpd.exe,1234\n"
	_ = parseWmicOutput(output)
}

func TestCaptureLogs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	orch := NewOrchestrator(ctx)
	orch.SetRuntime(&mockRuntime{})

	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte("test log\n"))
		pw.Close()
	}()

	orch.captureLogs("Test", pr)
}

func TestOrchestrator_Additional(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	orch := NewOrchestrator(ctx)
	orch.SetRuntime(&mockRuntime{})

	t.Run("shouldRefresh", func(t *testing.T) {
		orch.RequestRefresh()
		if !orch.shouldRefresh([]string{"Apache"}) {
			t.Error("Expected true after RequestRefresh")
		}

		orch.updateCache("Apache", ServiceDetailedInfo{Status: "Running", PID: 0})
		if !orch.shouldRefresh([]string{"Apache"}) {
			t.Error("Expected true for running service with PID 0")
		}
	})

	t.Run("updateOpenSSLInfo", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		info := &ServiceDetailedInfo{}
		orch.updateOpenSSLInfo(info, tempDir)

		sslDir := filepath.Join(tempDir, "ssl")
		os.MkdirAll(sslDir, 0755)
		os.WriteFile(filepath.Join(sslDir, "ca.crt"), []byte("fake cert"), 0644)
		orch.updateOpenSSLInfo(info, tempDir)
		if info.Status != "Running" {
			t.Errorf("Expected Running for OpenSSL, got %s", info.Status)
		}
	})

	t.Run("extractOsteniaPID", func(t *testing.T) {
		pid := extractOsteniaPID("C:\\bin\\apache\\httpd.exe", "1234", "C:\\bin")
		if pid != 1234 {
			t.Errorf("Expected 1234, got %d", pid)
		}
		pid = extractOsteniaPID("C:\\other\\httpd.exe", "1234", "C:\\bin")
		if pid != 0 {
			t.Errorf("Expected 0 for non-ostenia path, got %d", pid)
		}
	})

	t.Run("updateNodeAndPythonInfo", func(t *testing.T) {
		origIsSystemPath := IsPathInSystemPath
		defer func() { IsPathInSystemPath = origIsSystemPath }()
		IsPathInSystemPath = func(path string) bool { return true }

		info := &ServiceDetailedInfo{}
		orch.updateNodeInfo(info, t.TempDir())
		if info.Status != "Running" {
			t.Errorf("Expected Running for Node.js, got %s", info.Status)
		}

		info = &ServiceDetailedInfo{}
		orch.updatePythonInfo(info, t.TempDir())
		if info.Status != "Running" {
			t.Errorf("Expected Running for Python, got %s", info.Status)
		}
	})

	t.Run("resolveServiceVersion", func(t *testing.T) {
		tempDir := t.TempDir()
		info := &ServiceDetailedInfo{}
		orch.resolveServiceVersion(info, "PHP", filepath.Join(tempDir, "nonexistent"))
		if info.ActiveVersion != "" {
			t.Errorf("Expected empty version for non-existent path, got %s", info.ActiveVersion)
		}

		phpDir := filepath.Join(tempDir, "php-8.2.0")
		os.MkdirAll(phpDir, 0755)
		// We can't easily mock GetPHPVersion here because it's a regular function in php.go
		// but we can test the symlink branch
		targetDir := filepath.Join(tempDir, "target-v1")
		os.MkdirAll(targetDir, 0755)
		linkPath := filepath.Join(tempDir, "link")
		// On non-windows, we can use os.Symlink
		err := os.Symlink(targetDir, linkPath)
		if err == nil {
			info = &ServiceDetailedInfo{}
			orch.resolveServiceVersion(info, "Other", linkPath)
			if info.ActiveVersion != "target-v1" {
				t.Errorf("Expected target-v1, got %s", info.ActiveVersion)
			}
		}
	})

	t.Run("containsInt", func(t *testing.T) {
		if !containsInt([]int{1, 2, 3}, 2) {
			t.Error("Expected true")
		}
		if containsInt([]int{1, 2, 3}, 4) {
			t.Error("Expected false")
		}
	})

	t.Run("StopService_Error", func(t *testing.T) {
		err := orch.StopService("NonExistent")
		if err != nil {
			t.Errorf("StopService should not return error for non-existent service, got %v", err)
		}
	})
}
