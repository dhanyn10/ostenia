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
	"sort"
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
	Ports         []int  `json:"ports"`
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

	info := ServiceDetailedInfo{Name: name, Status: "Stopped", Ports: []int{}}

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
	case "Apache": exeName = "httpd.exe"
	case "MySQL":  exeName = "mysqld.exe"
	case "HeidiSQL": exeName = "heidisql.exe"
	case "Nginx":  exeName = "nginx.exe"
	case "PHP":    exeName = "php-cgi.exe"
	}

	if exeName == "" { return info }

	pids := findPIDsByName(exeName)
	if len(pids) > 0 {
		info.Status = "Running"
		info.PID = pids[0]

		if name == "Apache" || name == "Nginx" || name == "MySQL" || name == "PHP" {
			// Find all ports used by ANY of the discovered PIDs
			info.Ports = findPortsForPIDs(pids)

			if len(info.Ports) > 0 {
				info.Port = info.Ports[0]
			} else if tracked {
				info.Port = s.port
				if info.Port > 0 {
					info.Ports = append(info.Ports, info.Port)
				}
			}
		}

		if !tracked || (info.Port > 0 && s.port != info.Port) {
			o.mu.Lock()
			o.services[name] = &runningService{cmd: nil, port: info.Port}
			o.mu.Unlock()
		}
	} else {
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
	if runtime.GOOS != "windows" { return pids }
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+exeName, "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil { return pids }
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.Contains(line, exeName) { continue }
		parts := strings.Split(line, ",")
		if len(parts) > 1 {
			pidStr := strings.Trim(parts[1], "\"")
			pid, _ := strconv.Atoi(pidStr)
			if pid > 0 { pids = append(pids, pid) }
		}
	}
	return pids
}

// findPortsForPIDs runs netstat once and collects all listening ports for the given PIDs
func findPortsForPIDs(pids []int) []int {
	ports := []int{}
	if len(pids) == 0 { return ports }

	pidMap := make(map[string]bool)
	for _, pid := range pids {
		pidMap[strconv.Itoa(pid)] = true
	}

	cmd := exec.Command("cmd", "/c", "netstat -ano")
	out, err := cmd.Output()
	if err != nil { return ports }

	lines := strings.Split(string(out), "\n")
	foundPorts := make(map[int]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "LISTENING") {
			continue
		}

		fields := strings.Fields(line)
		// Expected: Proto, Local Addr, Foreign Addr, State, PID
		if len(fields) >= 5 {
			pidStr := fields[len(fields)-1]
			if pidMap[pidStr] {
				localAddr := fields[1]
				lastColon := strings.LastIndex(localAddr, ":")
				if lastColon != -1 {
					p, err := strconv.Atoi(localAddr[lastColon+1:])
					if err == nil && p > 0 {
						foundPorts[p] = true
					}
				}
			}
		}
	}

	for p := range foundPorts {
		ports = append(ports, p)
	}
	sort.Ints(ports)
	return ports
}

func (o *Orchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
	o.mu.Lock()
	if _, exists := o.services[name]; exists {
		s := o.services[name]
		if s.cmd != nil && s.cmd.Process != nil {
			pids := findPIDsByName(filepath.Base(binaryPath))
			isStillRunning := false
			for _, p := range pids {
				if p == s.cmd.Process.Pid { isStillRunning = true; break }
			}
			if isStillRunning {
				o.mu.Unlock()
				return fmt.Errorf("service %s is already running", name)
			}
		}
	}
	o.mu.Unlock()

	absPath, _ := filepath.Abs(binaryPath)
	cmd := exec.Command(absPath, args...)
	if workingDir != "" { cmd.Dir = workingDir }
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go o.captureLogs(name, stdout)
	go o.captureLogs(name, stderr)

	if err := cmd.Start(); err != nil { return err }

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
		case "Apache": exeNames = []string{"httpd.exe"}
		case "MySQL":  exeNames = []string{"mysqld.exe"}
		case "HeidiSQL": exeNames = []string{"heidisql.exe"}
		case "Nginx":  exeNames = []string{"nginx.exe"}
		case "PHP":    exeNames = []string{"php.exe", "php-cgi.exe"}
		}
		for _, exe := range exeNames {
			exec.Command("taskkill", "/F", "/IM", exe, "/T").Run()
		}
		time.Sleep(600 * time.Millisecond)
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
		wruntime.EventsEmit(o.ctx, "service_log", map[string]string{"service": name, "message": scanner.Text()})
	}
}

func (o *Orchestrator) emitStatus(name string, status string) {
	info := o.GetDetailedInfo(name)
	wruntime.EventsEmit(o.ctx, "service_status", info)
}

func (o *Orchestrator) StopAll() {
	o.mu.Lock()
	names := make([]string, 0, len(o.services))
	for name := range o.services { names = append(names, name) }
	o.mu.Unlock()
	for _, name := range names { o.StopService(name) }
}
