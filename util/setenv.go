//+build windows

package util

import (
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	varName = "SSH_AUTH_SOCK"
	wslName = "WSLENV"
)

func notifySystem() {
	var (
		mod             = syscall.NewLazyDLL("user32")
		proc            = mod.NewProc("SendMessageW")
		wmSETTINGCHANGE = uint32(0x001A)
	)
	_, _, _ = proc.Call(uintptr(windows.InvalidHandle), uintptr(wmSETTINGCHANGE), 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))))
}

// PrepareUserEnvironment modifies user environment. It sets SSH_AUTH_SOCK and creates/changes WSLENV to make sure path is
// available to all WSL environments started after this fuction was called.
func PrepareUserEnvironment(path string, debug bool) error {

	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.READ|registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.SetStringValue(varName, path); err != nil {
		return err
	} else if debug {
		log.Printf("Set '%s=%s'", varName, path)
	}

	val, _, err := k.GetStringValue(wslName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if debug {
		log.Printf("Was '%s=%s'", wslName, val)
	}

	parts := strings.Split(val, ":")
	vals := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(parts) == 0 {
			continue
		}
		if !strings.HasPrefix(part, varName) {
			vals = append(vals, part)
		}
	}
	vals = append(vals, varName+"/up")
	val = strings.Join(vals, ":")

	if err := k.SetStringValue(wslName, val); err != nil {
		return err
	} else if debug {
		log.Printf("Set '%s=%s'", wslName, val)
	}

	notifySystem()
	return nil
}

// CleanUserEnvironment will reverse settings done by PrepareUserEnvironment.
func CleanUserEnvironment(debug bool) error {

	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.READ|registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()

	if err := k.DeleteValue(varName); err != nil {
		return err
	} else if debug {
		log.Printf("Del '%s'", varName)
	}

	val, _, err := k.GetStringValue(wslName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if debug {
		log.Printf("Was '%s=%s'", wslName, val)
	}

	parts := strings.Split(val, ":")
	vals := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(parts) == 0 {
			continue
		}
		if !strings.HasPrefix(part, varName) {
			vals = append(vals, part)
		}
	}
	val = strings.Join(vals, ":")

	if err := k.SetStringValue(wslName, val); err != nil {
		return err
	} else if debug {
		log.Printf("Set '%s=%s'", wslName, val)
	}

	notifySystem()
	return nil
}
