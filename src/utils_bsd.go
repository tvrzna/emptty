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
