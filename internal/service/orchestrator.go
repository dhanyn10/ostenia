package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
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

// ServiceDetailedInfo contains comprehensive status and metadata for a service
type ServiceDetailedInfo struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	PID           int    `json:"pid"`
	Port          int    `json:"port"`
	Ports         []int  `json:"ports"`
	RemainingDays int    `json:"remainingDays,omitempty"`
	ActiveVersion string `json:"activeVersion,omitempty"`
}

type runningService struct {
	cmd  *exec.Cmd
	port int
}

// Orchestrator manages the lifecycle, monitoring, and state of background services
type Orchestrator struct {
	ctx          context.Context
	services     map[string]*runningService
	serviceCache map[string]ServiceDetailedInfo
	mu           sync.Mutex
	activeTab    string
	tabMu        sync.RWMutex
	needsRefresh bool
	refreshMu    sync.Mutex
}

// NewOrchestrator creates a new Orchestrator instance
func NewOrchestrator(ctx context.Context) *Orchestrator {
	return &Orchestrator{
		ctx:          ctx,
		services:     make(map[string]*runningService),
		serviceCache: make(map[string]ServiceDetailedInfo),
		activeTab:    "activity",
		needsRefresh: true,
	}
}

// SetActiveTab updates the current view state and requests a refresh if navigating to activity
func (o *Orchestrator) SetActiveTab(tab string) {
	o.tabMu.Lock()
	oldTab := o.activeTab
	o.activeTab = tab
	o.tabMu.Unlock()
	if oldTab == "plugins" && tab == "activity" {
		o.RequestRefresh()
	}
}

// RequestRefresh marks the orchestrator state as needing a UI update
func (o *Orchestrator) RequestRefresh() {
	o.refreshMu.Lock()
	o.needsRefresh = true
	o.refreshMu.Unlock()
}

// StartWatcher begins a background ticker that periodically updates service statuses
func (o *Orchestrator) StartWatcher() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		servicesToWatch := []string{"Apache", "Nginx", "MySQL", "PHP", "Node.js", "Python", "HeidiSQL", "OpenSSL"}
		for {
			select {
			case <-o.ctx.Done():
				return
			case <-ticker.C:
				o.handleWatcherTick(servicesToWatch)
			}
		}
	}()
}

func (o *Orchestrator) handleWatcherTick(servicesToWatch []string) {
	o.tabMu.RLock()
	currentTab := o.activeTab
	o.tabMu.RUnlock()

	if currentTab != "activity" {
		return
	}

	if o.shouldRefresh(servicesToWatch) {
		for _, name := range servicesToWatch {
			info := o.updateServiceInfo(name)
			wruntime.EventsEmit(o.ctx, "service_status", info)
		}
	}
}

func (o *Orchestrator) shouldRefresh(servicesToWatch []string) bool {
	o.refreshMu.Lock()
	if o.needsRefresh {
		o.needsRefresh = false
		o.refreshMu.Unlock()
		return true
	}
	o.refreshMu.Unlock()

	o.mu.Lock()
	defer o.mu.Unlock()
	for _, name := range servicesToWatch {
		if o.isServiceMissingCriticalInfo(name) {
			return true
		}
	}
	return false
}

func (o *Orchestrator) isServiceMissingCriticalInfo(name string) bool {
	cached := o.serviceCache[name]
	if cached.Status == "Running" && name != "OpenSSL" && name != "Node.js" && name != "Python" {
		if cached.PID == 0 || len(cached.Ports) == 0 {
			return true
		}
	}
	return false
}

// IsRunning returns true if a service is marked as Running in the cache
func (o *Orchestrator) IsRunning(name string) bool {
	o.mu.Lock()
	info, ok := o.serviceCache[name]
	o.mu.Unlock()
	if ok {
		return info.Status == "Running"
	}
	return false
}

// GetDetailedInfo retrieves the cached status for a service
func (o *Orchestrator) GetDetailedInfo(name string) ServiceDetailedInfo {
	o.mu.Lock()
	defer o.mu.Unlock()
	if info, ok := o.serviceCache[name]; ok {
		return info
	}
	return ServiceDetailedInfo{Name: name, Status: "Stopped", Ports: []int{}}
}

func (o *Orchestrator) updateServiceInfo(name string) ServiceDetailedInfo {
	o.mu.Lock()
	s, tracked := o.services[name]
	o.mu.Unlock()

	info := ServiceDetailedInfo{Name: name, Status: "Stopped", Ports: []int{}}
	baseDir := config.GetBaseDir()

	switch name {
	case "Node.js":
		o.updateNodeInfo(&info, baseDir)
	case "Python":
		o.updatePythonInfo(&info, baseDir)
	case "OpenSSL":
		o.updateOpenSSLInfo(&info, baseDir)
	case "HeidiSQL":
		o.updateHeidiSQLInfo(&info)
	case "PHP", "Apache", "MySQL", "Nginx":
		o.updateGenericServiceInfo(&info, name, baseDir, s, tracked)
	}

	o.updateCache(name, info)
	return info
}

func (o *Orchestrator) updateNodeInfo(info *ServiceDetailedInfo, baseDir string) {
	currentPath := filepath.Join(baseDir, "bin", "nodejs", "current")
	if IsPathInSystemPath(currentPath) {
		info.Status = "Running"
	}
	ver, err := GetNodeVersion(currentPath)
	if err == nil {
		info.ActiveVersion = ver
	}
}

func (o *Orchestrator) updatePythonInfo(info *ServiceDetailedInfo, baseDir string) {
	currentPath := filepath.Join(baseDir, "bin", "python", "current")
	if IsPathInSystemPath(currentPath) {
		info.Status = "Running"
	}
	ver, err := GetPythonVersion(currentPath)
	if err == nil {
		info.ActiveVersion = ver
	}
}

func (o *Orchestrator) updateOpenSSLInfo(info *ServiceDetailedInfo, baseDir string) {
	caPath := filepath.Join(baseDir, "ssl", "ca.crt")
	if _, err := os.Stat(caPath); err == nil {
		info.Status = "Running"
		days, _ := ssl.GetRemainingDays(caPath)
		info.RemainingDays = days
	}
}

func (o *Orchestrator) updateHeidiSQLInfo(info *ServiceDetailedInfo) {
	exePath, _ := utils.DetectHeidiSQLInstallation()
	if exePath != "" {
		info.Status = "Running"
		info.ActiveVersion = "System"
	}
}

func (o *Orchestrator) updateGenericServiceInfo(info *ServiceDetailedInfo, name string, baseDir string, s *runningService, tracked bool) {
	currentPath := filepath.Join(baseDir, "bin", strings.ToLower(name), "current")
	if name == "PHP" {
		if IsPathInUserPath(currentPath) {
			info.Status = "Running"
		}
		ver, err := GetPHPVersion(currentPath)
		if err == nil {
			info.ActiveVersion = ver
		}
	} else if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		info.ActiveVersion = filepath.Base(resolved)
	}

	exeMap := map[string]string{
		"Apache": "httpd.exe",
		"MySQL":  "mysqld.exe",
		"Nginx":  "nginx.exe",
		"PHP":    "php-cgi.exe",
	}

	exeName := exeMap[name]
	pids := findOsteniaPIDs(exeName)
	if len(pids) > 0 {
		info.Status = "Running"
		info.PID = pids[0]
		o.updateServicePorts(info, pids, s, tracked)
		if !tracked || (info.Port > 0 && s.port != info.Port) {
			o.mu.Lock()
			o.services[name] = &runningService{cmd: nil, port: info.Port}
			o.mu.Unlock()
		}
	} else if tracked {
		if name != "PHP" || !IsPathInUserPath(filepath.Join(baseDir, "bin", "php", "current")) {
			o.mu.Lock()
			delete(o.services, name)
			o.mu.Unlock()
		}
	}
}

func (o *Orchestrator) updateServicePorts(info *ServiceDetailedInfo, pids []int, s *runningService, tracked bool) {
	foundPorts := make(map[int]bool)
	for _, pid := range pids {
		ports := findPortsByPIDExact(pid)
		for _, p := range ports {
			foundPorts[p] = true
		}
	}
	for p := range foundPorts {
		info.Ports = append(info.Ports, p)
	}
	sort.Ints(info.Ports)
	if len(info.Ports) > 0 {
		info.Port = info.Ports[0]
	} else if tracked && s.port > 0 {
		info.Port = s.port
		info.Ports = append(info.Ports, info.Port)
	}
}

func (o *Orchestrator) updateCache(name string, info ServiceDetailedInfo) {
	o.mu.Lock()
	o.serviceCache[name] = info
	o.mu.Unlock()
}

func findOsteniaPIDs(exeName string) []int {
	pids := []int{}
	if runtime.GOOS != "windows" {
		return pids
	}
	baseDir := config.GetBaseDir()
	binPath := filepath.Join(baseDir, "bin")
	wmicPath := filepath.Join(utils.GetSystemDirectory(), "wbem", "wmic.exe")
	cmd := exec.Command(wmicPath, "process", "where", fmt.Sprintf("name='%s'", exeName), "get", "ExecutablePath,ProcessId", "/format:csv")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return pids
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if pid := parseWmicLine(line, binPath); pid > 0 {
			pids = append(pids, pid)
		}
	}
	return pids
}

func parseWmicLine(line, binPath string) int {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(strings.ToLower(line), "node") {
		return 0
	}
	parts := strings.Split(line, ",")
	if len(parts) >= 3 {
		execPath := strings.TrimSpace(parts[1])
		pidStr := strings.TrimSpace(parts[2])
		if strings.HasPrefix(strings.ToLower(execPath), strings.ToLower(binPath)) {
			pid, err := strconv.Atoi(pidStr)
			if err == nil && pid > 0 {
				return pid
			}
		}
	}
	return 0
}

func findPortsByPIDExact(pid int) []int {
	ports := []int{}
	if pid <= 0 {
		return ports
	}
	pidStr := strconv.Itoa(pid)

	netstatPath := filepath.Join(utils.GetSystemDirectory(), "netstat.exe")
	cmd := exec.Command(netstatPath, "-ano")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ports
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if p := parseNetstatLine(line, pidStr); p > 0 {
			if !contains(ports, p) {
				ports = append(ports, p)
			}
		}
	}
	return ports
}

func parseNetstatLine(line, pidStr string) int {
	line = strings.TrimSpace(line)
	if line == "" || !strings.Contains(line, "LISTENING") || !strings.Contains(line, pidStr) {
		return 0
	}
	fields := strings.Fields(line)
	if len(fields) >= 5 && fields[len(fields)-1] == pidStr && fields[3] == "LISTENING" {
		localAddr := fields[1]
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon != -1 {
			p, err := strconv.Atoi(localAddr[lastColon+1:])
			if err == nil && p > 0 {
				return p
			}
		}
	}
	return 0
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// StartServiceWithPort launches a service and tracks it by its known port
func (o *Orchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
	if o.isAlreadyRunning(name, binaryPath) {
		return fmt.Errorf("service %s is already running", name)
	}

	absPath, _ := filepath.Abs(binaryPath)
	cmd := exec.Command(absPath, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	utils.SetHideWindow(cmd)

	o.setupPipesAndStart(name, cmd, port)
	return nil
}

func (o *Orchestrator) isAlreadyRunning(name, binaryPath string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if s, exists := o.services[name]; exists {
		if s.cmd != nil && s.cmd.Process != nil {
			pids := findOsteniaPIDs(filepath.Base(binaryPath))
			for _, p := range pids {
				if p == s.cmd.Process.Pid {
					return true
				}
			}
		}
	}
	return false
}

func (o *Orchestrator) setupPipesAndStart(name string, cmd *exec.Cmd, port int) {
	stdout, err := cmd.StdoutPipe()
	if err == nil {
		go o.captureLogs(name, stdout)
	}
	stderr, err := cmd.StderrPipe()
	if err == nil {
		go o.captureLogs(name, stderr)
	}
	if err := cmd.Start(); err != nil {
		return
	}
	o.mu.Lock()
	o.services[name] = &runningService{cmd: cmd, port: port}
	o.mu.Unlock()
	o.RequestRefresh()
	o.emitStatus(name, "Running")
	go o.waitForProcessExit(name, cmd)
}

func (o *Orchestrator) waitForProcessExit(name string, cmd *exec.Cmd) {
	_ = cmd.Wait()
	time.Sleep(500 * time.Millisecond)
	o.mu.Lock()
	delete(o.services, name)
	o.mu.Unlock()
	o.RequestRefresh()
	o.emitStatus(name, "Stopped")
}

// StartService launches a service without a specific port requirement
func (o *Orchestrator) StartService(name string, binaryPath string, args []string, workingDir string) error {
	return o.StartServiceWithPort(name, binaryPath, args, workingDir, 0)
}

// StopService gracefully or forcefully shuts down a running service
func (o *Orchestrator) StopService(name string) error {
	if runtime.GOOS == "windows" {
		if name == "HeidiSQL" {
			o.stopHeidiSQL()
		} else {
			o.stopGenericService(name)
		}
	}
	o.mu.Lock()
	delete(o.services, name)
	o.mu.Unlock()
	o.RequestRefresh()
	o.emitStatus(name, "Stopped")
	return nil
}

func (o *Orchestrator) stopHeidiSQL() {
	_, uninstaller := utils.DetectHeidiSQLInstallation()
	if uninstaller != "" {
		cmdPath := filepath.Join(utils.GetSystemDirectory(), "cmd.exe")
		cmd := exec.Command(cmdPath, "/c", "start", "", uninstaller)
		cmd.Env = utils.SafeEnv()
		utils.SetHideWindow(cmd)
		_ = cmd.Run()
	} else {
		o.killProcessesByExe("heidisql.exe")
	}
}

func (o *Orchestrator) stopGenericService(name string) {
	exeNames := getExesForService(name)
	for _, exe := range exeNames {
		o.killProcessesByExe(exe)
	}
	time.Sleep(600 * time.Millisecond)
}

func getExesForService(name string) []string {
	switch name {
	case "Apache":
		return []string{"httpd.exe"}
	case "MySQL":
		return []string{"mysqld.exe"}
	case "HeidiSQL":
		return []string{"heidisql.exe"}
	case "Nginx":
		return []string{"nginx.exe"}
	case "PHP":
		return []string{"php.exe", "php-cgi.exe"}
	case "Node.js":
		return []string{"node.exe"}
	case "Python":
		return []string{"python.exe"}
	default:
		return nil
	}
}

func (o *Orchestrator) killProcessesByExe(exe string) {
	pids := findOsteniaPIDs(exe)
	taskkillPath := filepath.Join(utils.GetSystemDirectory(), "taskkill.exe")
	for _, pid := range pids {
		killCmd := exec.Command(taskkillPath, "/F", "/PID", strconv.Itoa(pid), "/T")
		killCmd.Env = utils.SafeEnv()
		utils.SetHideWindow(killCmd)
		_ = killCmd.Run()
	}
}

func (o *Orchestrator) captureLogs(name string, reader io.ReadCloser) {
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		wruntime.EventsEmit(o.ctx, "service_log", map[string]string{"service": name, "message": scanner.Text()})
	}
}

func (o *Orchestrator) emitStatus(name string, status string) {
	info := o.updateServiceInfo(name)
	wruntime.EventsEmit(o.ctx, "service_status", info)
}

// StopAll terminates all services currently being tracked by the orchestrator
func (o *Orchestrator) StopAll() {
	o.mu.Lock()
	names := make([]string, 0, len(o.services))
	for name := range o.services {
		names = append(names, name)
	}
	o.mu.Unlock()
	for _, name := range names {
		_ = o.StopService(name)
	}
}
