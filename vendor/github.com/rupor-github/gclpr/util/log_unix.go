//go:build linux || darwin

package util

import (
	"io"
	"log"
)

// NewLogWriter redirects all log output depending on debug parameetr.
// When true all output goes to OutputDebugString and you could use debugger or Sysinternals dbgview.exe to collect it.
// When false - everything is discarded.
func NewLogWriter(title string, flags int, debug bool) {

	log.SetPrefix("[" + title + "] ")
	log.SetFlags(flags)

	if !debug {
		log.SetOutput(io.Discard)
	}
}
