//go:build dragonfly || freebsd || netbsd || openbsd

package src

import (
	"os"
	"syscall"
)

const (
	currentVc = "/dev/tty"
)

// Sets fsuid, fsgid and fsgroups according sysuser
func setFsUser(usr *sysuser) {
	handleErr(syscall.Setuid(usr.uid))

	handleErr(syscall.Setgid(usr.gid))
}

// Sets keyboard LEDs
func setKeyboardLeds(tty *os.File, scrolllock, numlock, capslock bool) {
	// Not implemented yet
}

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
