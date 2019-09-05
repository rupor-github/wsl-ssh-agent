//+build windows

package util

import (
	"syscall"
	"unsafe"
)

// MsgType specifies how message box will look.
type MsgType uint32

// Actual values
const (
	MsgError       MsgType = 0x00000010
	MsgExclamation MsgType = 0x00000030
	MsgInformation MsgType = 0x00000040
)

// ShowOKMessage shows MB_OK message box.
func ShowOKMessage(t MsgType, title, text string) {
	var (
		mod  = syscall.NewLazyDLL("user32")
		proc = mod.NewProc("MessageBoxW")
		mb   = 0x00000000 + t
	)
	_, _, _ = proc.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(mb))
}
