//go:build nopam
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
	if conf.Autologin && conf.DefaultUser != "" {
		usr, err := user.Lookup(conf.DefaultUser)
		handleErr(err)
		return getSysuser(usr)
	}
	hostname, _ := os.Hostname()
	var username string
	if conf.DefaultUser != "" {
		if !conf.HideEnterLogin {
			fmt.Printf("%s login: %s\n", hostname, conf.DefaultUser)
		}
		username = conf.DefaultUser
	} else {
		if !conf.HideEnterLogin {
			fmt.Printf("%s login: ", hostname)
		}
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		handleErr(err)
		username = input[:len(input)-1]
	}
	if !conf.HideEnterPassword {
		fmt.Print("Password: ")
	}
	password, err := readPassword()
	handleErr(err)

	if authPassword(username, password) {
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

// Opens auth session with XDG_SESSION_TYPE set directly into PAM environments
func openAuthSession(sessionType string) error {
	// Nothing to do here
	return nil
}
