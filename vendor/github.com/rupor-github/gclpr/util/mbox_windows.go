//go:build windows

package util

import (
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32   = windows.NewLazySystemDLL("user32")
	pMessageBox = modUser32.NewProc("MessageBoxW")
)

// Windows SDK constants.
const (
	// How it will look.
	MB_OK                = 0x00000000
	MB_OKCANCEL          = 0x00000001
	MB_ABORTRETRYIGNORE  = 0x00000002
	MB_YESNOCANCEL       = 0x00000003
	MB_YESNO             = 0x00000004
	MB_RETRYCANCEL       = 0x00000005
	MB_CANCELTRYCONTINUE = 0x00000006
	MB_ICONHAND          = 0x00000010
	MB_ICONQUESTION      = 0x00000020
	MB_ICONEXCLAMATION   = 0x00000030
	MB_ICONASTERISK      = 0x00000040
	MB_USERICON          = 0x00000080
	MB_ICONWARNING       = MB_ICONEXCLAMATION
	MB_ICONERROR         = MB_ICONHAND
	MB_ICONINFORMATION   = MB_ICONASTERISK
	MB_ICONSTOP          = MB_ICONHAND
	// And behave.
	MB_DEFBUTTON1    = 0x00000000
	MB_DEFBUTTON2    = 0x00000100
	MB_DEFBUTTON3    = 0x00000200
	MB_DEFBUTTON4    = 0x00000300
	MB_SETFOREGROUND = 0x00010000
	// Return values.
	IDOK     = 1
	IDCANCEL = 2
	IDABORT  = 3
	IDRETRY  = 4
	IDIGNORE = 5
	IDYES    = 6
	IDNO     = 7
)

// MessageBox full native implementation.
func MessageBox(title, text string, style uintptr) int {
	pText, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return -1
	}
	pTitle, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return -1
	}
	ret, _, _ := pMessageBox.Call(0,
		uintptr(unsafe.Pointer(pText)),
		uintptr(unsafe.Pointer(pTitle)),
		style)
	return int(ret)
}

// MsgType specifies how message box will look and behave.
type MsgType uint32

// Actual values.
const (
	MsgError       MsgType = MB_ICONHAND
	MsgExclamation MsgType = MB_ICONEXCLAMATION
	MsgInformation MsgType = MB_ICONASTERISK
)

// ShowOKMessage shows simple MB_OK message box.
func ShowOKMessage(t MsgType, title, text string) {

	log.Print(text)

	_, _, _ = pMessageBox.Call(0,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(title))),
		uintptr(t))
}
