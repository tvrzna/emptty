// +build nopam

package main

// #include <crypt.h>
// #include <shadow.h>
// #include <string.h>
// #include <stdlib.h>
// #cgo linux LDFLAGS: -lcrypt
import "C"

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"unsafe"
)

// Handle authentication of user without PAM.
// If user is successfully authorized, it returns sysuser.
//
// If autologin is enabled, it behaves as user has been authorized.
func authUser(conf *config) *sysuser {
	if conf.autologin && conf.defaultUser != "" {
		usr, err := user.Lookup(conf.defaultUser)
		handleErr(err)
		return getSysuser(usr)
	}
	hostname, _ := os.Hostname()
	var username string
	if conf.defaultUser != "" {
		fmt.Printf("%s login: %s\n", hostname, conf.defaultUser)
		username = conf.defaultUser
	} else {
		fmt.Printf("%s login: ", hostname)
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		handleErr(err)
		username = input[:len(input)-1]
	}
	fmt.Print("Password: ")
	password, err := readPassword()
	handleErr(err)

	if authPassword(username, password) {
		password = ""
		usr, err := user.Lookup(username)
		username = ""

		handleErr(err)

		return getSysuser(usr)
	}
	handleStrErr("Authentication failure")
	return nil
}

// Tries to authorize user with password.
func authPassword(username string, password string) bool {
	usr := C.CString(username)
	defer C.free(unsafe.Pointer(usr))

	passwd := C.CString(password)
	defer C.free(unsafe.Pointer(passwd))

	pwd := C.getspnam(usr)
	if pwd == nil {
		return false
	}
	crypted := C.crypt(passwd, pwd.sp_pwdp)
	if C.strcmp(crypted, pwd.sp_pwdp) != 0 {
		return false
	}
	return true
}

// Handles close of authentication
func closeAuth() {
	// Nothing to do here
}

// Defines specific environmental variables defined by PAM
func defineSpecificEnvVariables() {
	// Nothing to do here
}
