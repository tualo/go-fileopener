//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func showError(err error) {
	const (
		mbOK        = 0x00000000
		mbIconError = 0x00000010
		mbTopmost   = 0x00040000
	)
	message, messageErr := syscall.UTF16PtrFromString(err.Error())
	title, titleErr := syscall.UTF16PtrFromString("Tualo File Opener")
	if messageErr != nil || titleErr != nil {
		return
	}
	messageBox := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	_, _, _ = messageBox.Call(
		0,
		uintptr(unsafe.Pointer(message)),
		uintptr(unsafe.Pointer(title)),
		uintptr(mbOK|mbIconError|mbTopmost),
	)
}
