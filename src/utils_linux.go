package src

import (
	"syscall"
)

// Sets fsuid, fsgid and fsgroups according sysuser
func setFsUser(usr *sysuser) {
	err := syscall.Setfsuid(usr.uid)
	handleErr(err)

	err = syscall.Setfsgid(usr.gid)
	handleErr(err)

	err = syscall.Setfsgid(usr.gid)
	handleErr(err)
}
