//go:build windows
// +build windows

package util

import (
	"log"
	"os"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	wslName = "WSLENV"
)

func notifySystem() {
	var (
		mod             = windows.NewLazySystemDLL("user32")
		proc            = mod.NewProc("SendMessageTimeoutW")
		wmSETTINGCHANGE = uint32(0x001A)
		smtoABORTIFHUNG = uint32(0x0002)
		smtoNORMAL      = uint32(0x0000)
	)

	start := time.Now()
	log.Printf("Broadcasting environment change. From %s", start)

	_, _, _ = proc.Call(uintptr(windows.InvalidHandle),
		uintptr(wmSETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Environment"))),
		uintptr(smtoNORMAL|smtoABORTIFHUNG),
		uintptr(1000),
		0)

	log.Printf("Broadcasting environment change. To   %s, Elapsed %s", time.Now(), time.Since(start))
}

// PrepareUserEnvironment modifies user environment. It sets SSH_AUTH_SOCK and creates/changes WSLENV to make sure path is
// available to all WSL environments started after this fuction was called.
func PrepareUserEnvironment(name, path string) error {

	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.READ|registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue(name, path); err != nil {
		return err
	}
	log.Printf("Set '%s=%s'", name, path)

	val, _, err := k.GetStringValue(wslName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	log.Printf("Was '%s=%s'", wslName, val)

	parts := strings.Split(val, ":")
	vals := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		if !strings.HasPrefix(part, name) {
			vals = append(vals, part)
		}
	}
	vals = append(vals, name+"/up")
	val = strings.Join(vals, ":")

	if err := k.SetStringValue(wslName, val); err != nil {
		return err
	}
	log.Printf("Set '%s=%s'", wslName, val)

	notifySystem()
	return nil
}

// CleanUserEnvironment will reverse settings done by PrepareUserEnvironment.
func CleanUserEnvironment(name string) error {

	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.READ|registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.DeleteValue(name); err != nil {
		return err
	}
	log.Printf("Del '%s'", name)

	val, _, err := k.GetStringValue(wslName)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	log.Printf("Was '%s=%s'", wslName, val)

	parts := strings.Split(val, ":")
	vals := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		if !strings.HasPrefix(part, name) {
			vals = append(vals, part)
		}
	}
	val = strings.Join(vals, ":")

	if len(val) == 0 {
		if err := k.DeleteValue(wslName); err != nil {
			return err
		}
		log.Printf("Del '%s'", wslName)
	} else {
		if err := k.SetStringValue(wslName, val); err != nil {
			return err
		}
		log.Printf("Set '%s=%s'", wslName, val)
	}

	notifySystem()
	return nil
}
