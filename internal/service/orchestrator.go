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
	"ostenia/internal/backend/interfaces"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ServiceDetailedInfo contains comprehensive status and metadata for a service
type ServiceDetailedInfo = interfaces.ServiceDetailedInfo

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
	runtime      interfaces.Runtime
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

// SetRuntime sets the Wails runtime for events emission
func (o *Orchestrator) SetRuntime(r interfaces.Runtime) {
	o.runtime = r
}

func (o *Orchestrator) emitEvent(eventName string, optionalData ...interface{}) {
	if o.runtime != nil && o.ctx != nil {
		o.runtime.EventsEmit(o.ctx, eventName, optionalData...)
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
				o.performWatcherCheck(servicesToWatch)
			}
		}
	}()
}

func (o *Orchestrator) performWatcherCheck(services []string) {
	o.tabMu.RLock()
	currentTab := o.activeTab
	o.tabMu.RUnlock()

	if currentTab != "activity" {
		return
	}

	if o.shouldRefresh(services) {
		for _, name := range services {
			info := o.updateServiceInfo(name)
			o.emitEvent("service_status", info)
		}
	}
}

func (o *Orchestrator) shouldRefresh(services []string) bool {
	o.refreshMu.Lock()
	if o.needsRefresh {
		o.needsRefresh = false
		o.refreshMu.Unlock()
		return true
	}
	o.refreshMu.Unlock()

	o.mu.Lock()
	defer o.mu.Unlock()
	for _, name := range services {
		cached := o.serviceCache[name]
		if cached.Status == "Running" && name != "OpenSSL" && name != "Node.js" && name != "Python" {
			if cached.PID == 0 || len(cached.Ports) == 0 {
				return true
			}
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
	o.resolveServiceVersion(info, name, currentPath)

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
		o.ensureServiceTracked(name, info, s, tracked)
	} else if tracked {
		o.cleanupUntrackedService(name, baseDir)
	}
}

func (o *Orchestrator) resolveServiceVersion(info *ServiceDetailedInfo, name, currentPath string) {
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
}

func (o *Orchestrator) ensureServiceTracked(name string, info *ServiceDetailedInfo, s *runningService, tracked bool) {
	if !tracked || (info.Port > 0 && s.port != info.Port) {
		o.mu.Lock()
		o.services[name] = &runningService{cmd: nil, port: info.Port}
		o.mu.Unlock()
	}
}

func (o *Orchestrator) cleanupUntrackedService(name, baseDir string) {
	if name != "PHP" || !IsPathInUserPath(filepath.Join(baseDir, "bin", "php", "current")) {
		o.mu.Lock()
		delete(o.services, name)
		o.mu.Unlock()
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


func parseWmicOutput(output string) []int {
	pids := []int{}
	binPath := filepath.Join(config.GetBaseDir(), "bin")
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "node") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			if pid := extractOsteniaPID(parts[1], parts[2], binPath); pid > 0 {
				pids = append(pids, pid)
			}
		}
	}
	return pids
}

func extractOsteniaPID(execPath, pidStr, binPath string) int {
	execPath = strings.TrimSpace(execPath)
	if strings.HasPrefix(strings.ToLower(execPath), strings.ToLower(binPath)) {
		pid, _ := strconv.Atoi(strings.TrimSpace(pidStr))
		return pid
	}
	return 0
}

var findPortsByPIDExactOverride func(pid int) []int

func findPortsByPIDExact(pid int) []int {
	if findPortsByPIDExactOverride != nil {
		return findPortsByPIDExactOverride(pid)
	}
	if pid <= 0 {
		return []int{}
	}

	netstatPath := filepath.Join(utils.GetSystemDirectory(), "netstat.exe")
	cmd := utils.Executor.Command(netstatPath, "-ano")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return []int{}
	}

	return parseNetstatOutput(string(out), pid)
}

func parseNetstatOutput(output string, pid int) []int {
	pidStr := strconv.Itoa(pid)
	ports := []int{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "LISTENING") || !strings.Contains(line, pidStr) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[len(fields)-1] == pidStr && fields[3] == "LISTENING" {
			if p := extractPortFromNetstatLine(fields[1]); p > 0 {
				if !containsInt(ports, p) {
					ports = append(ports, p)
				}
			}
		}
	}
	return ports
}

func extractPortFromNetstatLine(localAddr string) int {
	lastColon := strings.LastIndex(localAddr, ":")
	if lastColon != -1 {
		p, _ := strconv.Atoi(localAddr[lastColon+1:])
		return p
	}
	return 0
}

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// StartServiceWithPort launches a service and tracks it by its known port
func (o *Orchestrator) StartServiceWithPort(name string, binaryPath string, args []string, workingDir string, port int) error {
	if o.isServiceAlreadyRunning(name, binaryPath) {
		return fmt.Errorf("service %s is already running", name)
	}

	cmd, err := o.setupServiceCommand(name, binaryPath, args, workingDir)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	o.mu.Lock()
	o.services[name] = &runningService{cmd: cmd, port: port}
	o.mu.Unlock()
	o.RequestRefresh()
	o.emitStatus(name, "Running")

	go o.waitForServiceExit(name, cmd)
	return nil
}

func (o *Orchestrator) isServiceAlreadyRunning(name, binaryPath string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	s, exists := o.services[name]
	if !exists || s.cmd == nil || s.cmd.Process == nil {
		return false
	}

	pids := findOsteniaPIDs(filepath.Base(binaryPath))
	for _, p := range pids {
		if p == s.cmd.Process.Pid {
			return true
		}
	}
	return false
}

func (o *Orchestrator) setupServiceCommand(name, binaryPath string, args []string, workingDir string) (*exec.Cmd, error) {
	absPath, _ := filepath.Abs(binaryPath)
	cmd := utils.Executor.Command(absPath, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	utils.SetHideWindow(cmd)

	if stdout, err := cmd.StdoutPipe(); err == nil {
		go o.captureLogs(name, stdout)
	}
	if stderr, err := cmd.StderrPipe(); err == nil {
		go o.captureLogs(name, stderr)
	}

	return cmd, nil
}

func (o *Orchestrator) waitForServiceExit(name string, cmd *exec.Cmd) {
	_ = cmd.Wait()
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
	o.mu.Lock()
	s, exists := o.services[name]
	o.mu.Unlock()

	if exists && s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}

	if runtime.GOOS == "windows" {
		o.stopServiceWindows(name)
		time.Sleep(600 * time.Millisecond)
	}

	o.mu.Lock()
	delete(o.services, name)
	o.mu.Unlock()
	o.RequestRefresh()
	o.emitStatus(name, "Stopped")
	return nil
}

func (o *Orchestrator) stopServiceWindows(name string) {
	if name == "HeidiSQL" {
		o.stopHeidiSQLWindows()
		return
	}

	exeMap := map[string][]string{
		"Apache":   {"httpd.exe"},
		"MySQL":    {"mysqld.exe"},
		"Nginx":    {"nginx.exe"},
		"PHP":      {"php.exe", "php-cgi.exe"},
		"Node.js":  {"node.exe"},
		"Python":   {"python.exe"},
	}

	taskkillPath := filepath.Join(utils.GetSystemDirectory(), "taskkill.exe")
	for _, exe := range exeMap[name] {
		pids := findOsteniaPIDs(exe)
		for _, pid := range pids {
			killCmd := utils.Executor.Command(taskkillPath, "/F", "/PID", strconv.Itoa(pid), "/T")
			killCmd.Env = utils.SafeEnv()
			utils.SetHideWindow(killCmd)
			_ = killCmd.Run()
		}
	}
}

func (o *Orchestrator) stopHeidiSQLWindows() {
	_, uninstaller := utils.DetectHeidiSQLInstallation()
	if uninstaller != "" {
		cmdPath := filepath.Join(utils.GetSystemDirectory(), "cmd.exe")
		cmd := utils.Executor.Command(cmdPath, "/c", "start", "", uninstaller)
		cmd.Env = utils.SafeEnv()
		utils.SetHideWindow(cmd)
		_ = cmd.Run()
	} else {
		taskkillPath := filepath.Join(utils.GetSystemDirectory(), "taskkill.exe")
		pids := findOsteniaPIDs("heidisql.exe")
		for _, pid := range pids {
			killCmd := utils.Executor.Command(taskkillPath, "/F", "/PID", strconv.Itoa(pid), "/T")
			killCmd.Env = utils.SafeEnv()
			utils.SetHideWindow(killCmd)
			_ = killCmd.Run()
		}
	}
}

func (o *Orchestrator) captureLogs(name string, reader io.ReadCloser) {
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		o.emitEvent("service_log", map[string]string{"service": name, "message": scanner.Text()})
	}
}

func (o *Orchestrator) emitStatus(name string, status string) {
	info := o.updateServiceInfo(name)
	o.emitEvent("service_status", info)
}

var findOsteniaPIDsOverride func(exeName string) []int

func findOsteniaPIDs(exeName string) []int {
	if findOsteniaPIDsOverride != nil {
		return findOsteniaPIDsOverride(exeName)
	}
	if runtime.GOOS != "windows" {
		return []int{}
	}

	wmicPath := filepath.Join(utils.GetSystemDirectory(), "wbem", "wmic.exe")
	cmd := utils.Executor.Command(wmicPath, "process", "where", fmt.Sprintf("name='%s'", exeName), "get", "ExecutablePath,ProcessId", "/format:csv")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return []int{}
	}

	return parseWmicOutput(string(out))
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
