//go:build windows

package util

import (
	"io"
	"log"
	"unsafe"

	"golang.org/x/sys/windows"
)

var kernel = windows.NewLazySystemDLL("kernel32")

// DebugWriter redirects all output to OutputDebugString().
type logWriter struct {
	proc *windows.LazyProc
}

// NewLogWriter redirects all log output depending on debug parameetr.
// When true all output goes to OutputDebugString and you could use debugger or Sysinternals dbgview.exe to collect it.
// When false - everything is discarded.
func NewLogWriter(title string, flags int, debug bool) {

	log.SetPrefix("[" + title + "] ")
	log.SetFlags(flags)

	if debug {
		res := &logWriter{proc: kernel.NewProc("OutputDebugStringW")}
		log.SetOutput(res)
	} else {
		log.SetOutput(io.Discard)
	}
}

func (l *logWriter) Write(p []byte) (n int, err error) {

	text, err := windows.UTF16PtrFromString(string(p))
	if err != nil {
		return 0, err
	}
	_, _, _ = l.proc.Call(uintptr(unsafe.Pointer(text)))
	return len(p), nil
}
