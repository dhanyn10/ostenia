//go:build windows

package network

import (
	"path/filepath"

	"golang.org/x/sys/windows"
	"ostenia/internal/plugins/utils"
)

func findPortOwnerOS(targetPort int) (int, string, string, error) {
	netstatPath := filepath.Join(utils.GetSystemDirectory(), "netstat.exe")
	cmd := utils.Executor.Command(netstatPath, "-ano")
	cmd.Env = utils.SafeEnv()
	utils.SetHideWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return 0, "", "", err
	}

	pid, err := parsePortOwnerFromNetstat(string(out), targetPort)
	if err != nil {
		return 0, "", "", err
	}

	execPath := getProcessPathByPID(pid)
	procName := ""
	if execPath != "" {
		procName = filepath.Base(execPath)
	} else if pid == 4 {
		procName = "System (HTTP.sys / IIS)"
	} else {
		procName = "Unknown Process"
	}

	return pid, procName, execPath, nil
}

func getProcessPathByPID(pid int) string {
	if pid <= 0 {
		return ""
	}
	hProc, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(hProc)

	buf := make([]uint16, 1024)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(hProc, 0, &buf[0], &size); err == nil {
		return windows.UTF16ToString(buf[:size])
	}
	return ""
}
