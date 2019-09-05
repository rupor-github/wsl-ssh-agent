package util

import (
	"syscall"
)

// Shared names.
const (
	AgentPipeName  = "\\\\.\\pipe\\openssh-ssh-agent"
	MaxNameLen     = syscall.UNIX_PATH_MAX
	MaxAgentMsgLen = 256 * 1024 // same as in openssh-portable
)
