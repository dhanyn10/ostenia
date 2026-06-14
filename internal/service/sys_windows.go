//go:build windows

package service

import (
	"os/exec"
	"syscall"
	"unsafe"
)

func notifyEnvironmentUpdate() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessage := user32.NewProc("SendMessageTimeoutW")
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)
	envStr, _ := syscall.UTF16PtrFromString("Environment")
	sendMessage.Call(uintptr(HWND_BROADCAST), uintptr(WM_SETTINGCHANGE), 0, uintptr(unsafe.Pointer(envStr)), uintptr(SMTO_ABORTIFHUNG), uintptr(5000), 0)
}
