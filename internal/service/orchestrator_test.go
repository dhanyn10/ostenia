package service

import (
	"context"
	"io"
	"os"
	"runtime"
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

func newTestOrchestrator() (context.Context, context.CancelFunc, *Orchestrator) {
	ctx, cancel := context.WithCancel(context.Background())
	orch := NewOrchestrator()
	orch.SetRuntime(&mockRuntime{})
	return ctx, cancel, orch
}

func TestOrchestrator_TabManagement(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

	orch.SetActiveTab("plugins")
	orch.SetActiveTab("activity")
	orch.RequestRefresh()
}

func TestOrchestrator_ServiceStatus(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

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
}

func TestOrchestrator_UpdateServiceInfoMocks(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

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

	orch.updateServiceInfo(serviceNodeJS)
	orch.updateServiceInfo("MySQL")
	orch.updateServiceInfo("Nginx")
	orch.updateServiceInfo("PHP")
	orch.updateServiceInfo("Python")
	orch.updateServiceInfo("HeidiSQL")
	orch.updateServiceInfo("OpenSSL")
}

func TestOrchestrator_StartStopService(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &testutil.MockExecutor{Output: ""}

	ctx, cancel, orch := newTestOrchestrator()
	defer cancel()

	binary := "true"
	if runtime.GOOS == "windows" {
		binary = "cmd"
	}
	err := orch.StartService(ctx, "TestService", binary, []string{}, "")
	if err != nil {
		t.Errorf("StartService failed: %v", err)
	}

	orch.StopAll(ctx)
	time.Sleep(100 * time.Millisecond)
}

func TestOrchestrator_WatcherAndRefresh(t *testing.T) {
	ctx, cancel, orch := newTestOrchestrator()
	defer cancel()

	orch.RequestRefresh()
	orch.performWatcherCheck(ctx, []string{"Apache"})

	ctx2, cancel2 := context.WithCancel(context.Background())
	orch2 := NewOrchestrator()
	orch2.StartWatcher(ctx2)
	time.Sleep(100 * time.Millisecond)
	cancel2()
}

func TestOrchestrator_ServiceTrackingAndCleanup(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

	info := &ServiceDetailedInfo{Name: "Apache", Status: "Running", Port: 80}
	orch.ensureServiceTracked("Apache", info, nil, false)
	orch.cleanupUntrackedService("Apache", "/tmp")
}

func TestOrchestrator_NetstatAndPorts(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

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
}

func TestOrchestrator_StopService_Windows(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

	orch.stopServiceWindows("Apache")
	orch.stopServiceWindows("HeidiSQL")
}

func TestOrchestrator_ResolveServiceVersion(t *testing.T) {
	_, cancel, orch := newTestOrchestrator()
	defer cancel()

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "v1.0.0")
	os.MkdirAll(target, 0755)

	link := filepath.Join(tempDir, "current")
	if runtime.GOOS != "windows" {
		os.Symlink(target, link)
	} else {
		link = target
	}

	info := &ServiceDetailedInfo{}
	orch.resolveServiceVersion(info, "apache", link)
}

func TestOrchestrator_ContainsInt(t *testing.T) {
	if !containsInt([]int{1, 2, 3}, 2) {
		t.Error("Expected true")
	}
	if containsInt([]int{1, 2, 3}, 4) {
		t.Error("Expected false")
	}
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
	orch := NewOrchestrator()
	orch.SetRuntime(&mockRuntime{})

	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte("test log\n"))
		pw.Close()
	}()

	orch.captureLogs(ctx, "Test", pr)
}

func TestOrchestrator_Additional(t *testing.T) {
	orch := NewOrchestrator()
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
}
