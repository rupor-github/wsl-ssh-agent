//go:build windows

package util

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	advapi32   = windows.NewLazySystemDLL("advapi32.dll")
	procGetAce = advapi32.NewProc("GetAce")
)

// Acl - for whatever reason golang.org/x/sys/windows package does not export field names on ACL and does not provide methods to do anything.
type Acl struct {
	AclRevision uint8
	Sbz1        uint8
	AclSize     uint16
	AceCount    uint16
	Sbz2        uint16
}

// Windows ACE_TYPE constants
const (
	ACCESS_ALLOWED_ACE_TYPE = 0
	ACCESS_DENIED_ACE_TYPE  = 1
)

// AccessAllowedAce is not defined in golang.org/x/sys/windows (yet?)
type AccessAllowedAce struct {
	AceType    uint8
	AceFlags   uint8
	AceSize    uint16
	AccessMask uint32
	SidStart   uint32
}

func getAce(acl *Acl, index uint32, ace **AccessAllowedAce) error {
	ret, _, _ := procGetAce.Call(uintptr(unsafe.Pointer(acl)), uintptr(index), uintptr(unsafe.Pointer(ace)))
	if int(ret) != 0 {
		return windows.GetLastError()
	}
	return nil
}

func getSid(name string) (*windows.SID, error) {

	if len(name) == 0 {
		info, err := windows.GetCurrentProcessToken().GetTokenUser()
		if err != nil {
			return nil, err
		}
		return info.User.Sid.Copy()
	}

	sid, _, _, err := windows.LookupSID("", name)
	if err != nil {
		return nil, err
	}
	return sid, nil
}

func checkPermissions(fname string, readOK bool) error {

	fi, err := os.Stat(fname)
	if err != nil || !fi.Mode().IsRegular() {
		return fmt.Errorf("not a regular file %s", fname)
	}

	// Ref: https://github.com/PowerShell/openssh-portable/blob/latestw_all/contrib/win32/win32compat/w32-sshfileperm.c

	var userSid, tiSid *windows.SID
	if userSid, err = getSid(""); err != nil {
		return fmt.Errorf("failed to retrieve the user sid: %w", err)
	}
	if tiSid, err = getSid("NT SERVICE\\TrustedInstaller"); err != nil {
		return fmt.Errorf("failed to retrieve the TI sid: %w", err)
	}

	sd, err := windows.GetNamedSecurityInfo(fname, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return fmt.Errorf("failed to retrieve security descriptor (owner sid and dacl) of file %s: %w", fname, err)
	}
	ownerSid, _, err := sd.Owner()
	if err != nil {
		return fmt.Errorf("failed to retrieve owner sid of file %s: %w", fname, err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return fmt.Errorf("failed to retrieve dacl of file %s: %w", fname, err)
	}
	if !ownerSid.IsValid() {
		return fmt.Errorf("invalid owner sid of file %s", fname)
	}

	if !ownerSid.IsWellKnown(windows.WinBuiltinAdministratorsSid) &&
		!ownerSid.IsWellKnown(windows.WinLocalSystemSid) &&
		!ownerSid.Equals(userSid) &&
		!(tiSid != nil && ownerSid.Equals(tiSid)) {
		return fmt.Errorf("bad owner of file %s", fname)
	}

	// iterate all aces of the file to find out if there is violation of the following rules:
	//     1. no others than administrators group, system account and current user account have write permission on the file

	ddacl := (*Acl)(unsafe.Pointer(dacl))
	for i := uint32(0); i < uint32(ddacl.AceCount); i++ {

		var currentAce *AccessAllowedAce
		if err := getAce(ddacl, i, &currentAce); err != nil {
			return fmt.Errorf("failed getAce %w", err)
		}

		// only interested in Allow ACE
		if currentAce.AceType != ACCESS_ALLOWED_ACE_TYPE {
			continue
		}

		currentTrusteeSid := (*windows.SID)(unsafe.Pointer(&currentAce.SidStart))
		currentAccessMask := currentAce.AccessMask

		// no need to check administrators group, user account, and system account
		if currentTrusteeSid.IsWellKnown(windows.WinBuiltinAdministratorsSid) ||
			currentTrusteeSid.IsWellKnown(windows.WinLocalSystemSid) ||
			currentTrusteeSid.Equals(userSid) ||
			(tiSid != nil && currentTrusteeSid.Equals(tiSid)) {
			continue
		}

		const (
			FILE_WRITE_DATA       = 0x0002
			FILE_APPEND_DATA      = 0x0004
			FILE_WRITE_EA         = 0x0010
			FILE_WRITE_ATTRIBUTES = 0x0100
		)

		// if read is allowed, allow ACES that do not give write access
		if readOK && (currentAccessMask&(FILE_WRITE_DATA|FILE_WRITE_ATTRIBUTES|FILE_WRITE_EA|FILE_APPEND_DATA)) == 0 {
			continue
		}

		// do reverse lookups on the sids to verify the sids are not actually for the same user as could be the case of a sidhistory entry in the ace
		resolvedUserAccount, resolvedUserDomain, ressolvedUserType, err1 := userSid.LookupAccount("")
		resolvedTrusteeAccount, resolvedTrusteeDomain, ressolvedTrusteeType, err2 := currentTrusteeSid.LookupAccount("")

		if err1 == nil && err2 == nil &&
			strings.EqualFold(resolvedUserAccount, resolvedTrusteeAccount) &&
			strings.EqualFold(resolvedUserDomain, resolvedTrusteeDomain) &&
			ressolvedUserType == ressolvedTrusteeType {
			continue
		}

		return fmt.Errorf("bad permissions - try removing permissions for user: %s\\%s (%s) on file %s", resolvedTrusteeDomain,
			resolvedTrusteeAccount, currentTrusteeSid, fname)
	}

	return nil
}
