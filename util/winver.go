//+build windows

package util

import (
	"log"
	"strconv"

	"golang.org/x/sys/windows/registry"
)

// IsProperWindowsVer checks if program is running on supported WIndows.
func IsProperWindowsVer() (bool, error) {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.READ)
	if err != nil {
		return false, err
	}
	defer k.Close()

	m, _, err := k.GetIntegerValue("CurrentMajorVersionNumber")
	if err != nil {
		return false, err
	}
	cb, _, err := k.GetStringValue("CurrentBuild")
	if err != nil {
		return false, err
	}
	b, err := strconv.Atoi(cb)
	if err != nil {
		return false, err
	}
	log.Printf("Windows - CurrentMajorVersionNumber: %d, CurrentBuild: %d", m, b)

	return m >= 10 && b >= 17063, nil
}
