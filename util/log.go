//+build windows

package util

import (
	"syscall"
	"unsafe"
)

// DebugWriter redirects all output to OutputDebugString().
type DebugWriter struct {
	dll  *syscall.LazyDLL
	proc *syscall.LazyProc
}

// NewDebugWriter allocates new logger instance.
func NewDebugWriter() *DebugWriter {
	res := &DebugWriter{
		dll: syscall.NewLazyDLL("kernel32"),
	}
	res.proc = res.dll.NewProc("OutputDebugStringW")
	return res
}

func (l *DebugWriter) Write(p []byte) (n int, err error) {
	text, err := syscall.UTF16PtrFromString(string(p))
	if err != nil {
		return 0, err
	}
	_, _, _ = l.proc.Call(uintptr(unsafe.Pointer(text)))
	return len(p), nil
}
