//go:build !nopam
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
	trans, _ = pam.StartFunc("emptty", conf.DefaultUser, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			if conf.Autologin {
				break
			}
			if conf.DefaultUser != "" && !conf.HideEnterLogin {
				hostname, _ := os.Hostname()
				fmt.Printf("%s login: %s\n", hostname, conf.DefaultUser)
			}
			if !conf.HideEnterPassword {
				fmt.Print("Password: ")
			}
			return readPassword()
		case pam.PromptEchoOn:
			if conf.Autologin {
				break
			}
			if !conf.HideEnterLogin {
				hostname, _ := os.Hostname()
				fmt.Printf("%s login: ", hostname)
			}
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
		return "", errors.New("unrecognized message style")
	})

	if err := trans.Authenticate(pam.Silent); err != nil {
		bkpErr := errors.New(err.Error())
		username, _ := trans.GetItem(pam.User)
		addBtmpEntry(username, os.Getpid(), conf.strTTY())
		handleErr(bkpErr)
	}
	logPrint("Authenticate OK")

	handleErr(trans.AcctMgmt(pam.Silent))
	handleErr(trans.SetItem(pam.Tty, "tty"+conf.strTTY()))
	handleErr(trans.SetCred(pam.EstablishCred))

	pamUsr, _ := trans.GetItem(pam.User)
	usr, _ := user.Lookup(pamUsr)

	return getSysuser(usr)
}

// Handles close of PAM authentication
func closeAuth() {
	if trans != nil {
		if err := trans.SetCred(pam.DeleteCred); err != nil {
			logPrint(err)
		}
		if err := trans.CloseSession(pam.Silent); err != nil {
			logPrint(err)
		}
		trans = nil
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

// Opens session with XDG_SESSION_TYPE set directly into PAM environments
func openSession(sessionType string) error {
	if trans != nil {
		if err := trans.PutEnv(fmt.Sprintf("XDG_SESSION_TYPE=%s", sessionType)); err != nil {
			return err
		}
		return trans.OpenSession(pam.Silent)
	}
	return errors.New("no active transaction")
}
