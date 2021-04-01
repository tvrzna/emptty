// +build dragonfly freebsd netbsd openbsd

package src

import (
	"syscall"
)

// Sets fsuid, fsgid and fsgroups according sysuser
func setFsUser(usr *sysuser) {
	err := syscall.Setuid(usr.uid)
	handleErr(err)

	err = syscall.Setgid(usr.gid)
	handleErr(err)
}

// Sets keyboard LEDs
func setKeyboardLeds(tty *os.File, scrolllock, numlock, capslock bool) {
	// Not implemented yet
}
