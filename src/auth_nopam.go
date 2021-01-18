// +build nopam

package src

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
)

const tagPam = "nopam"

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
	addBtmpEntry(username, os.Getpid(), conf.strTTY())
	handleStrErr("Authentication failure")
	return nil
}

// Handles close of authentication
func closeAuth() {
	// Nothing to do here
}

// Defines specific environmental variables defined by PAM
func defineSpecificEnvVariables(usr *sysuser) {
	// Nothing to do here
}
