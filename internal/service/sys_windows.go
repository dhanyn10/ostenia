//go:build windows

package service

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func notifyEnvironmentUpdate() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessage := user32.NewProc("SendMessageTimeoutW")
	const (
		hwndBroadcast   = 0xFFFF
		wmSettingChange = 0x001A
		smtoAbortIfHung = 0x0002
	)
	envStr, _ := syscall.UTF16PtrFromString("Environment")
	sendMessage.Call(uintptr(hwndBroadcast), uintptr(wmSettingChange), 0, uintptr(unsafe.Pointer(envStr)), uintptr(smtoAbortIfHung), uintptr(5000), 0)
}

func findOsteniaPIDsWindows(exeName, binPath string) []int {
	if binPath == "" {
		return []int{}
	}

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return []int{}
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	if err := windows.Process32First(snapshot, &entry); err != nil {
		return []int{}
	}

	var pids []int
	for {
		name := windows.UTF16ToString(entry.ExeFile[:])
		if strings.EqualFold(name, exeName) && isProcessInBinPath(entry.ProcessID, binPath) {
			pids = append(pids, int(entry.ProcessID))
		}

		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}
	return pids
}

func isProcessInBinPath(pid uint32, binPath string) bool {
	exePath, err := getProcessImagePath(pid)
	if err != nil {
		return false
	}
	return IsPathInside(exePath, binPath)
}

func getProcessImagePath(pid uint32) (string, error) {
	hProc, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(hProc)

	buf := make([]uint16, 1024)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(hProc, 0, &buf[0], &size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf[:size]), nil
}
