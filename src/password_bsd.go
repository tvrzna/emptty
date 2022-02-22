// +build dragonfly freebsd netbsd openbsd

package src

import (
	"syscall"
)

// Enables or disables echo depending on status
func setTerminalEcho(fd uintptr, status bool) error {
	flag := ""
	if !status {
		flag = "-"
	}
	pid, err := syscall.ForkExec("/bin/stty", []string{"stty", flag + "echo"}, &syscall.ProcAttr{Dir: "", Files: []uintptr{fd}})
	if err == nil {
		syscall.Wait4(pid, nil, 0, nil)
	}
	return err
}
