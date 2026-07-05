package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	"testing"
)

type mockRuntime struct {
	interfaces.Runtime
}

func (m *mockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {}

type mockExecutor struct {
	output []byte
}

func (m *mockExecutor) Command(name string, args ...string) *exec.Cmd {
	argList := []string{"-test.run=TestHelperProcess", "--"}
	argList = append(argList, args...)
	cmd := exec.Command(os.Args[0], argList...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "MOCK_OUTPUT="+string(m.output))
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("MOCK_OUTPUT") != "" {
		fmt.Fprint(os.Stdout, os.Getenv("MOCK_OUTPUT"))
	}
	os.Exit(0)
}

func TestOrchestrator_Complete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockSys := NewMockSystem()
	orch := NewOrchestrator(ctx, mockSys)
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
		mockSys.PIDs["httpd.exe"] = []int{1234}

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
		_ = orch.StartService("TestService", "dummy", []string{}, "")
		orch.StopAll()
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
		mockSys.Ports[1234] = []int{80}

		ports := mockSys.FindPortsByPID(1234)
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

func TestCaptureLogs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockSys := NewMockSystem()
	orch := NewOrchestrator(ctx, mockSys)
	orch.SetRuntime(&mockRuntime{})

	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte("test log\n"))
		pw.Close()
	}()

	orch.captureLogs("Test", pr)
}
