package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ServiceDetailedInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	PID    int    `json:"pid"`
	Port   int    `json:"port"`
}

type runningService struct {
	cmd  *exec.Cmd
	port int
}

type Orchestrator struct {
	ctx      context.Context
	services map[string]*runningService
	mu       sync.Mutex
}

func NewOrchestrator(ctx context.Context) *Orchestrator {
	return &Orchestrator{
		ctx:      ctx,
		services: make(map[string]*runningService),
	}
}

func (o *Orchestrator) IsRunning(name string) bool {
	info := o.GetDetailedInfo(name)
	return info.Status == "Running"
}

func (o *Orchestrator) GetDetailedInfo(name string) ServiceDetailedInfo {
	o.mu.Lock()
	s, tracked := o.services[name]
	o.mu.Unlock()

	info := ServiceDetailedInfo{Name: name, Status: "Stopped"}

	if tracked && s.cmd != nil && s.cmd.Process != nil {
		info.Status = "Running"
		info.PID = s.cmd.Process.Pid
		info.Port = s.port
		return info
	}

	if runtime.GOOS == "windows" {
		var exeName string
		switch name {
		case "Apache":
			exeName = "httpd.exe"
		case "MySQL":
			exeName = "mysqld.exe"
		case "HeidiSQL":
			exeName = "heidisql.exe"
		case "Nginx":
			exeName = "nginx.exe"
		}

		if exeName != "" {
			out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+exeName, "/FO", "CSV", "/NH").Output()
			if err == nil {
				line := string(out)
				if strings.Contains(line, exeName) {
					info.Status = "Running"
					parts := strings.Split(line, ",")
					if len(parts) > 1 {
						pidStr := strings.Trim(parts[1], "\"")
						pid, _ := strconv.Atoi(pidStr)
						info.PID = pid
					}
				}
			}
		}
	}

	return info
}

func (o *Orchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
	o.mu.Lock()
	if _, exists := o.services[name]; exists {
		o.mu.Unlock()
		return fmt.Errorf("service %s is already running", name)
	}
	o.mu.Unlock()

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

	o.mu.Lock()
	o.services[name] = &runningService{cmd: cmd, port: port}
	o.mu.Unlock()

	o.emitStatus(name, "Running")

	go func() {
		cmd.Wait()
		time.Sleep(500 * time.Millisecond) // Give system time to clean up
		o.mu.Lock()
		delete(o.services, name)
		o.mu.Unlock()
		o.emitStatus(name, "Stopped")
	}()

	return nil
}

func (o *Orchestrator) StartService(name string, binaryPath string, args []string, workingDir string) error {
	return o.StartServiceWithPort(name, binaryPath, args, workingDir, 0)
}

func (o *Orchestrator) StopService(name string) error {
	if runtime.GOOS == "windows" {
		var exeNames []string
		switch name {
		case "Apache":
			exeNames = []string{"httpd.exe"}
		case "MySQL":
			exeNames = []string{"mysqld.exe"}
		case "HeidiSQL":
			exeNames = []string{"heidisql.exe"}
		case "Nginx":
			exeNames = []string{"nginx.exe"}
		case "PHP":
			exeNames = []string{"php.exe", "php-cgi.exe"}
		}

		for _, exe := range exeNames {
			exec.Command("taskkill", "/F", "/IM", exe, "/T").Run()
		}
		time.Sleep(500 * time.Millisecond) // Wait for process to fully terminate
	}

	o.mu.Lock()
	delete(o.services, name)
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
	info := o.GetDetailedInfo(name)
	wruntime.EventsEmit(o.ctx, "service_status", info)
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
