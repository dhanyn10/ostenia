//go:build windows

package service

import (
	"syscall"
	"unsafe"
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
