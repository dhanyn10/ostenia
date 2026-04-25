package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

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

// IsRunning checks if the service is running (either tracked or in system)
func (o *Orchestrator) IsRunning(name string) bool {
	o.mu.Lock()
	_, tracked := o.services[name]
	o.mu.Unlock()

	if tracked {
		return true
	}

	// If not tracked by us, check the OS process list
	if runtime.GOOS == "windows" {
		var exeName string
		switch name {
		case "Apache":
			exeName = "httpd.exe"
		case "MySQL":
			exeName = "mysqld.exe"
		case "HeidiSQL":
			exeName = "heidisql.exe"
		}

		if exeName != "" {
			// tasklist /FI "IMAGENAME eq exeName"
			out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+exeName, "/NH").Output()
			if err == nil && strings.Contains(string(out), exeName) {
				return true
			}
		}
	}

	return false
}

func (o *Orchestrator) StartService(name string, binaryPath string, args []string, workingDir string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Double check if already running in system before starting
	if _, exists := o.services[name]; exists {
		return fmt.Errorf("service %s is already tracked as running", name)
	}

	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(absPath, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

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
	// 1. Broad cleanup on Windows for specific services
	if runtime.GOOS == "windows" {
		var exeNames []string
		switch name {
		case "Apache":
			exeNames = []string{"httpd.exe"}
		case "MySQL":
			exeNames = []string{"mysqld.exe"}
		case "HeidiSQL":
			exeNames = []string{"heidisql.exe"}
		case "PHP":
			exeNames = []string{"php.exe", "php-cgi.exe"}
		}

		for _, exe := range exeNames {
			exec.Command("taskkill", "/F", "/IM", exe, "/T").Run()
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 2. Clear from tracked services
	o.mu.Lock()
	cmd, exists := o.services[name]
	if exists && cmd != nil && cmd.Process != nil {
		if runtime.GOOS != "windows" {
			cmd.Process.Kill()
		}
		delete(o.services, name)
	}
	o.mu.Unlock()

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
