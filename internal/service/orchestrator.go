package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"ostenia/internal/ssl"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ServiceDetailedInfo struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	Port          int    `json:"port"`
	RemainingDays int    `json:"remainingDays,omitempty"`
}

type runningService struct {
	cmd  *exec.Cmd
	port int
}

type Orchestrator struct {
	ctx       context.Context
	services  map[string]*runningService
	mu        sync.Mutex
	activeTab string
	tabMu     sync.RWMutex
}

func NewOrchestrator(ctx context.Context) *Orchestrator {
	return &Orchestrator{
		ctx:       ctx,
		services:  make(map[string]*runningService),
		activeTab: "activity",
	}
}

func (o *Orchestrator) SetActiveTab(tab string) {
	o.tabMu.Lock()
	defer o.tabMu.Unlock()
	o.activeTab = tab
	fmt.Printf("[Orchestrator] Active tab changed to: %s\n", tab)
}

func (o *Orchestrator) StartWatcher() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		servicesToWatch := []string{"Apache", "Nginx", "MySQL", "PHP", "HeidiSQL", "OpenSSL"}

		for {
			select {
			case <-o.ctx.Done():
				return
			case <-ticker.C:
				o.tabMu.RLock()
				currentTab := o.activeTab
				o.tabMu.RUnlock()

				if currentTab == "activity" {
					for _, name := range servicesToWatch {
						info := o.GetDetailedInfo(name)
						wruntime.EventsEmit(o.ctx, "service_status", info)
					}
				}
			}
		}
	}()
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

	// Special case for OpenSSL status based on Root CA existence
	if name == "OpenSSL" {
		baseDir := config.GetBaseDir()
		caPath := filepath.Join(baseDir, "ssl", "ca.crt")
		if _, err := os.Stat(caPath); err == nil {
			info.Status = "Running"
			days, _ := ssl.GetRemainingDays(caPath)
			info.RemainingDays = days
		}
		return info
	}

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
	case "PHP":
		exeName = "php-cgi.exe"
	}

	if exeName == "" {
		return info
	}

	// 1. Get all PIDs for this executable using CMD tasklist
	pids := findPIDsByName(exeName)

	if len(pids) > 0 {
		info.Status = "Running"
		// Default PID is the first one found
		info.PID = pids[0]

		// 2. If this is a service that needs a port, search for it via netstat
		if name == "Apache" || name == "Nginx" || name == "MySQL" || name == "PHP" {
			// Check each PID because web servers often have multiple processes (master/worker)
			for _, pid := range pids {
				port := findPortByPID(pid)
				if port > 0 {
					info.Port = port
					info.PID = pid // Use the PID that is actually holding the port
					break
				}
			}
		}

		// 3. Update internal state if not tracked or port has changed
		if !tracked || s.port != info.Port {
			o.mu.Lock()
			o.services[name] = &runningService{cmd: nil, port: info.Port}
			o.mu.Unlock()
		}
	} else {
		// If no PID found but we previously considered it running, remove from track
		if tracked {
			o.mu.Lock()
			delete(o.services, name)
			o.mu.Unlock()
		}
	}

	return info
}

func findPIDsByName(exeName string) []int {
	pids := []int{}
	if runtime.GOOS != "windows" {
		return pids
	}

	// Run cmd tasklist to find all PIDs for exeName
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+exeName, "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return pids
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.Contains(line, exeName) {
			continue
		}
		// tasklist /FO CSV returns like: "nginx.exe","1234","Console","1","10.000 K"
		parts := strings.Split(line, ",")
		if len(parts) > 1 {
			pidStr := strings.Trim(parts[1], "\"")
			pid, err := strconv.Atoi(pidStr)
			if err == nil && pid > 0 {
				pids = append(pids, pid)
			}
		}
	}
	return pids
}

func findPortByPID(pid int) int {
	if pid <= 0 {
		return 0
	}

	// Run cmd netstat -ano to find the port being LISTENED to by a specific PID
	cmd := exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano | findstr LISTENING | findstr %d", pid))
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(out), "\n")
	pidStr := strconv.Itoa(pid)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		// Expected fields: Proto, Local Addr, Foreign Addr, State, PID
		// Example: TCP    0.0.0.0:80             0.0.0.0:0              LISTENING       1234
		if len(fields) >= 5 && fields[len(fields)-1] == pidStr {
			localAddr := fields[1]
			// Local Addr is like 0.0.0.0:80 or [::]:80
			lastColon := strings.LastIndex(localAddr, ":")
			if lastColon != -1 {
				portStr := localAddr[lastColon+1:]
				port, parseErr := strconv.Atoi(portStr)
				if parseErr == nil && port > 0 {
					return port
				}
			}
		}
	}
	return 0
}

func (o *Orchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
	o.mu.Lock()
	if _, exists := o.services[name]; exists {
		// If already exists but status is actually Stopped, allow it to proceed
		s := o.services[name]
		if s.cmd != nil && s.cmd.Process != nil {
			// Check if actually still running
			pids := findPIDsByName(filepath.Base(binaryPath))
			running := false
			for _, p := range pids {
				if p == s.cmd.Process.Pid {
					running = true
					break
				}
			}
			if running {
				o.mu.Unlock()
				return fmt.Errorf("service %s is already running", name)
			}
		}
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
		time.Sleep(500 * time.Millisecond)
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
		time.Sleep(500 * time.Millisecond)
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
