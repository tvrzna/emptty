//go:build !nopam

package src

import (
	"errors"
	"fmt"
	"os"
	"os/user"

	"github.com/msteinert/pam/v2"
)

const tagPam = ""

type pamState byte

const (
	pamInit pamState = iota
	pamAuthenticated
	pamCredsEstablished
	pamSessionOpened
	pamClosed
)

// PamHandle defines structure of handle specifically designed for using PAM authorization
type pamHandle struct {
	*authBase
	trans *pam.Transaction
	u     *sysuser
	pamState
}

// Creates authHandle and handles authorization
func auth(conf *config) *pamHandle {
	h := &pamHandle{authBase: &authBase{}}
	h.authUser(conf)
	return h
}

// Handle PAM authentication of user.
// If user is successfully authorized, it returns sysuser.
//
// If autologin is enabled, it behaves as user has been authorized.
func (h *pamHandle) authUser(conf *config) {
	username, err := h.selectUser(conf)
	handleErr(err)
	if h.command != "" {
		return
	}

	h.pamState = pamInit
	h.trans, _ = pam.StartFunc("emptty", username, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			if conf.Autologin {
				break
			}
			if !conf.HideEnterPassword {
				fmt.Print(conf.GetIndentString() + "Password: ")
			}
			return readPassword()
		case pam.PromptEchoOn:
			return "", nil
		case pam.ErrorMsg:
			logPrint(msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println(msg)
			return "", nil
		}
		return "", errors.New("unrecognized message style")
	})

	if err := h.trans.Authenticate(pam.DisallowNullAuthtok); err != nil {
		bkpErr := errors.New(err.Error())
		username, _ := h.trans.GetItem(pam.User)
		addBtmpEntry(username, os.Getpid(), conf.strTTY())
		h.handleErr(bkpErr)
	}
	h.pamState = pamAuthenticated
	logPrint("Authenticate OK")

	h.handleErr(h.trans.AcctMgmt(pam.Silent))
	h.handleErr(h.trans.SetItem(pam.Tty, "tty"+conf.strTTY()))
	h.handleErr(h.trans.SetCred(pam.EstablishCred))
	h.pamState = pamCredsEstablished

	pamUsr, _ := h.trans.GetItem(pam.User)
	usr, _ := user.Lookup(pamUsr)

	h.u = getSysuser(usr)
	h.saveLastSelectedUser(conf, pamUsr)
}

func (h *pamHandle) handleErr(err error) {
	if err != nil {
		h.closeAuth()
		handleErr(err)
	}
}

// Gets sysuser
func (h *pamHandle) usr() *sysuser {
	return h.u
}

// Handles close of PAM authentication
func (h *pamHandle) closeAuth() {
	if h != nil && h.trans != nil && h.pamState < pamClosed {
		logPrint("Closing PAM auth")

		defer func() {
			if err := h.trans.End(); err != nil {
				logPrint(err)
			}
			h.trans = nil
			h.u = nil
			h.pamState = pamClosed
		}()

		if h.pamState >= pamSessionOpened {
			if err := h.trans.CloseSession(pam.Silent); err != nil {
				logPrint(err)
			}
		}

		if h.pamState >= pamCredsEstablished {
			if err := h.trans.SetCred(pam.DeleteCred); err != nil {
				logPrint(err)
			}
		}
	}
}

// Defines specific environmental variables defined by PAM
func (h *pamHandle) defineSpecificEnvVariables() {
	if h.trans != nil && h.u != nil {
		envs, _ := h.trans.GetEnvList()
		for key, value := range envs {
			h.u.setenv(key, value)
		}
	}
}

// Opens auth session with XDG_SESSION_TYPE set directly into PAM environments
func (h *pamHandle) openAuthSession(sessionType string) error {
	if h.trans != nil {
		if err := h.trans.PutEnv(fmt.Sprintf("XDG_SESSION_TYPE=%s", sessionType)); err != nil {
			return err
		}
		if err := h.trans.OpenSession(pam.Silent); err != nil {
			return err
		} else {
			h.pamState = pamSessionOpened
			return nil
		}
	}
	return errors.New("no active transaction")
}
