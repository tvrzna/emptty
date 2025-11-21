//go:build nopam

package src

import (
	"fmt"
	"os"
	"os/user"
	"strings"
)

const tagPam = "nopam"

// PamHandle defines structure of handle specifically designed for not using PAM authorization
type nopamHandle struct {
	*authBase
	u *sysuser
}

// Creates authHandle and handles authorization
func auth(conf *config) *nopamHandle {
	h := &nopamHandle{authBase: &authBase{}}
	h.authUser(conf)
	return h
}

// Handle authentication of user without PAM.
// If user is successfully authorized, it returns sysuser.
//
// If autologin is enabled, it behaves as user has been authorized.
func (n *nopamHandle) authUser(conf *config) {
	if conf.Autologin && conf.DefaultUser != "" {
		usr, err := user.Lookup(conf.DefaultUser)
		handleErr(err)
		n.u = getSysuser(usr)
		return
	}

	username, err := n.selectUser(conf)
	handleErr(err)
	if n.command != "" {
		return
	}

	if !conf.HideEnterPassword {
		indent := ""
		if conf.IndentSelection > 0 {
			indent = strings.Repeat(" ", conf.IndentSelection)
		}
		fmt.Print(indent + "Password: ")
	}
	password, err := readPassword()
	handleErr(err)

	if n.authPassword(username, password) {
		n.saveLastSelectedUser(conf, username)
		usr, err := user.Lookup(username)
		username = ""

		handleErr(err)

		n.u = getSysuser(usr)
		return
	}
	addBtmpEntry(username, os.Getpid(), conf.strTTY())
	handleStrErr("Authentication failure")
}

// Gets sysuser
func (n *nopamHandle) usr() *sysuser {
	return n.u
}

// Handles close of authentication
func (n *nopamHandle) closeAuth() {
	// Nothing to do here
}

// Defines specific environmental variables defined by PAM
func (n *nopamHandle) defineSpecificEnvVariables() {
	// Nothing to do here
}

// Opens auth session with XDG_SESSION_TYPE set directly into PAM environments
func (n *nopamHandle) openAuthSession(sessionType string) error {
	// Nothing to do here
	return nil
}
