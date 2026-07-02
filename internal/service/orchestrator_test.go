package service

import (
	"context"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/backend/interfaces"
	"testing"
)

type mockRuntime struct {
	interfaces.Runtime
}

func (m *mockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {}

type mockExecutor struct {
	utils.CommandExecutor
	output []byte
}

func (m *mockExecutor) Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "MOCK_OUTPUT="+string(m.output))
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	output := os.Getenv("MOCK_OUTPUT")
	os.Stdout.WriteString(output)
	os.Exit(0)
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
	})

	t.Run("UpdateServiceInfoMocks", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv("OSTENIA_HOME", tempDir)
		defer os.Unsetenv("OSTENIA_HOME")

		m := &mockExecutor{
			output: []byte("Node,C:\\bin\\apache\\httpd.exe,1234\n"),
		}
		utils.Executor = m

		orch.updateServiceInfo("Apache")
		orch.updateServiceInfo("Node.js")
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
