package service

import (
	"fmt"
	"os"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sort"
)

type WindowsSystem struct {
	executor utils.CommandExecutor
}

func NewWindowsSystem(executor utils.CommandExecutor) *WindowsSystem {
	if executor == nil {
		executor = utils.Executor
	}
	return &WindowsSystem{executor: executor}
}

func (w *WindowsSystem) GetPath(target string) (string, error) {
	getCmd := w.executor.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("[Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::%s)", target))
	utils.SetHideWindow(getCmd)
	out, err := getCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get %s path: %w", target, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (w *WindowsSystem) SetPath(path string, target string) error {
	escapedPath := strings.ReplaceAll(path, "'", "''")

	if target == "Machine" && !w.IsAdmin() {
		scriptContent := fmt.Sprintf("[Environment]::SetEnvironmentVariable('Path', '%s', [EnvironmentVariableTarget]::Machine)", escapedPath)
		f, err := os.CreateTemp("", "ostenia_set_path_*.ps1")
		if err != nil {
			return fmt.Errorf("failed to create temp script: %w", err)
		}
		tmpScript := f.Name()
		if _, err := f.Write([]byte(scriptContent)); err != nil {
			f.Close()
			os.Remove(tmpScript)
			return fmt.Errorf("failed to write temp script: %w", err)
		}
		f.Close()
		defer os.Remove(tmpScript)

		args := fmt.Sprintf("-NoProfile -ExecutionPolicy Bypass -File \"%s\"", tmpScript)
		elevatedCmd := fmt.Sprintf("Start-Process powershell -ArgumentList '%s' -Verb RunAs -Wait", args)
		cmd := w.executor.Command("powershell", "-NoProfile", "-Command", elevatedCmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("UAC prompt denied: %w", err)
		}
	} else {
		script := fmt.Sprintf("[Environment]::SetEnvironmentVariable('Path', '%s', [EnvironmentVariableTarget]::%s)", escapedPath, target)
		cmd := w.executor.Command("powershell", "-NoProfile", "-Command", script)
		utils.SetHideWindow(cmd)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to set %s path: %w", target, err)
		}
	}
	w.NotifyEnvironmentUpdate()
	return nil
}

func (w *WindowsSystem) FindOsteniaPIDs(exeName string) []int {
	if runtime.GOOS != "windows" {
		return []int{}
	}

	wmicPath := filepath.Join(utils.GetSystemDirectory(), "wbem", "wmic.exe")
	cmd := w.executor.Command(wmicPath, "process", "where", fmt.Sprintf("name='%s'", exeName), "get", "ExecutablePath,ProcessId", "/format:csv")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return []int{}
	}

	return w.parseWmicOutput(string(out))
}

func (w *WindowsSystem) parseWmicOutput(output string) []int {
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
			if pid := w.extractOsteniaPID(parts[1], parts[2], binPath); pid > 0 {
				pids = append(pids, pid)
			}
		}
	}
	return pids
}

func (w *WindowsSystem) extractOsteniaPID(execPath, pidStr, binPath string) int {
	execPath = strings.TrimSpace(execPath)
	if strings.HasPrefix(strings.ToLower(execPath), strings.ToLower(binPath)) {
		pid, _ := strconv.Atoi(strings.TrimSpace(pidStr))
		return pid
	}
	return 0
}

func (w *WindowsSystem) FindPortsByPID(pid int) []int {
	if pid <= 0 {
		return []int{}
	}

	netstatPath := filepath.Join(utils.GetSystemDirectory(), "netstat.exe")
	cmd := w.executor.Command(netstatPath, "-ano")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return []int{}
	}

	return w.parseNetstatOutput(string(out), pid)
}

func (w *WindowsSystem) parseNetstatOutput(output string, pid int) []int {
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
			if p := w.extractPortFromNetstatLine(fields[1]); p > 0 {
				if !w.containsInt(ports, p) {
					ports = append(ports, p)
				}
			}
		}
	}
	sort.Ints(ports)
	return ports
}

func (w *WindowsSystem) extractPortFromNetstatLine(localAddr string) int {
	lastColon := strings.LastIndex(localAddr, ":")
	if lastColon != -1 {
		p, _ := strconv.Atoi(localAddr[lastColon+1:])
		return p
	}
	return 0
}

func (w *WindowsSystem) containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func (w *WindowsSystem) KillProcess(pid int) error {
	taskkillPath := filepath.Join(utils.GetSystemDirectory(), "taskkill.exe")
	killCmd := w.executor.Command(taskkillPath, "/F", "/PID", strconv.Itoa(pid), "/T")
	killCmd.Env = utils.SafeEnv()
	utils.SetHideWindow(killCmd)
	return killCmd.Run()
}

func (w *WindowsSystem) CreateSymlink(target, link string) error {
	if runtime.GOOS == "windows" {
		cmd := w.executor.Command("cmd", "/c", "mklink", "/J", link, target)
		return cmd.Run()
	}
	return os.Symlink(target, link)
}

func (w *WindowsSystem) IsAdmin() bool {
	return IsAdmin()
}

func (w *WindowsSystem) NotifyEnvironmentUpdate() {
	notifyEnvironmentUpdate()
}

var _ interfaces.System = (*WindowsSystem)(nil)
