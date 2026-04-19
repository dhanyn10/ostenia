package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ServiceInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "Running", "Stopped", "Starting", "Stopping"
	Version string `json:"version"`
}

type Orchestrator struct {
	ctx      context.Context
	services map[string]*exec.Cmd
	mu       sync.Mutex
}

func NewOrchestrator(ctx context.Context) *Orchestrator {
	return &Orchestrator{
		ctx:      ctx,
		services: make(map[string]*exec.Cmd),
	}
}

func (o *Orchestrator) StartService(name string, binaryPath string, args []string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.services[name]; exists {
		return fmt.Errorf("service %s is already running", name)
	}

	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return err
	}

	// Mocking for non-windows environment
	if runtime.GOOS != "windows" {
		fmt.Printf("[MOCK] Starting service %s: %s %v\n", name, absPath, args)
		// We'll just simulate a running process
		cmd := exec.Command("sleep", "1000")
		err = cmd.Start()
		if err != nil {
			return err
		}
		o.services[name] = cmd
		go o.emitStatus(name, "Running")
		return nil
	}

	cmd := exec.Command(absPath, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go o.captureLogs(name, stdout)
	go o.captureLogs(name, stderr)

	if err := cmd.Start(); err != nil {
		return err
	}

	o.services[name] = cmd
	o.emitStatus(name, "Running")

	go func() {
		cmd.Wait()
		o.mu.Lock()
		delete(o.services, name)
		o.mu.Unlock()
		o.emitStatus(name, "Stopped")
	}()

	return nil
}

func (o *Orchestrator) StopService(name string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	cmd, exists := o.services[name]
	if !exists {
		return nil
	}

	if runtime.GOOS == "windows" {
		// On windows, we might need taskkill for some processes like Apache
		kill := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", cmd.Process.Pid))
		kill.Run()
	} else {
		cmd.Process.Kill()
	}

	delete(o.services, name)
	o.emitStatus(name, "Stopped")
	return nil
}

func (o *Orchestrator) captureLogs(name string, reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		wruntime.EventsEmit(o.ctx, "service_log", map[string]string{
			"service": name,
			"message": scanner.Text(),
		})
	}
}

func (o *Orchestrator) emitStatus(name string, status string) {
	wruntime.EventsEmit(o.ctx, "service_status", map[string]string{
		"name":   name,
		"status": status,
	})
}

func (o *Orchestrator) StopAll() {
	o.mu.Lock()
	names := make([]string, 0, len(o.services))
	for name := range o.services {
		names = append(names, name)
	}
	o.mu.Unlock()

	for _, name := range names {
		o.StopService(name)
	}
}
