// +build !nopam

package src

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/user"

	"github.com/msteinert/pam"
)

const tagPam = ""

var trans *pam.Transaction

// Handle PAM authentication of user.
// If user is successfully authorized, it returns sysuser.
//
// If autologin is enabled, it behaves as user has been authorized.
func authUser(conf *config) *sysuser {
	var err error

	trans, err = pam.StartFunc("emptty", conf.defaultUser, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			if conf.autologin {
				break
			}
			if conf.defaultUser != "" {
				hostname, _ := os.Hostname()
				fmt.Printf("%s login: %s\n", hostname, conf.defaultUser)
			}
			fmt.Print("Password: ")
			return readPassword()
		case pam.PromptEchoOn:
			if conf.autologin {
				break
			}
			hostname, _ := os.Hostname()
			fmt.Printf("%s login: ", hostname)
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			return input[:len(input)-1], nil
		case pam.ErrorMsg:
			logPrint(msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println(msg)
			return "", nil
		}
		return "", errors.New("Unrecognized message style")
	})

	err = trans.Authenticate(pam.Silent)
	if err != nil {
		bkpErr := errors.New(err.Error())
		username, _ := trans.GetItem(pam.User)
		addBtmpEntry(username, os.Getpid(), conf.strTTY())
		handleErr(bkpErr)
	}
	logPrint("Authenticate OK")

	err = trans.AcctMgmt(pam.Silent)
	handleErr(err)

	err = trans.SetItem(pam.Tty, "tty"+conf.strTTY())
	handleErr(err)

	err = trans.OpenSession(pam.Silent)
	handleErr(err)

	pamUsr, _ := trans.GetItem(pam.User)
	usr, _ := user.Lookup(pamUsr)

	return getSysuser(usr)
}

// Handles close of PAM authentication
func closeAuth() {
	if trans != nil {
		err := trans.CloseSession(pam.Silent)
		trans = nil
		if err != nil {
			logPrint(err)
		}
	}
}

// Defines specific environmental variables defined by PAM
func defineSpecificEnvVariables(usr *sysuser) {
	if trans != nil {
		envs, _ := trans.GetEnvList()
		for key, value := range envs {
			usr.setenv(key, value)
		}
	}
}
