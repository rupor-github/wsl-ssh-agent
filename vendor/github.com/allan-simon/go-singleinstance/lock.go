package singleinstance

import (
	"io/ioutil"
	"strconv"
)

// If filename is a lock file, returns the PID of the process locking it
func GetLockFilePid(filename string) (pid int, err error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	pid, err = strconv.Atoi(string(contents))
	return
}
