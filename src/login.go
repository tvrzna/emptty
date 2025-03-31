package src

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	loginRetryFile = "/tmp/emptty/login-retry"
)

// AuthHandle interface defines handle for authorization
type authHandle interface {
	usr() *sysuser
	authUser(*config)
	closeAuth()
	defineSpecificEnvVariables()
	openAuthSession(string) error
	getCommand() string
}

// Login into graphical environment
func login(conf *config, h *sessionHandle) string {
	h.auth = auth(conf)
	if h.auth != nil && h.auth.getCommand() != "" {
		return h.auth.getCommand()
	}

	if err := handleLoginRetries(conf, h.auth.usr()); err != nil {
		h.auth.closeAuth()
		handleStrErr("Exceeded maximum number of allowed login retries in short period.")
		return ""
	}

	d := processDesktopSelection(h.auth.usr(), conf)

	runDisplayScript(conf.DisplayStartScript)

	if err := h.auth.openAuthSession(d.env.sessionType()); err != nil {
		h.auth.closeAuth()
		handleStrErr("No active transaction")
		return ""
	}

	h.session = createSession(h.auth, d, conf)
	h.session.start()

	h.auth.closeAuth()

	runDisplayScript(conf.DisplayStopScript)

	return ""
}

// Process whole desktop load, selection and last used save.
func processDesktopSelection(usr *sysuser, conf *config) *desktop {
	d, usrLang := loadUserDesktop(usr.homedir)

	if d == nil || d.selection != SelectionFalse {
		selectedDesktop, lastDesktop := selectDesktop(usr, conf, d)
		if isLastDesktopForSave(usr, lastDesktop, selectedDesktop) {
			setUserLastSession(usr, selectedDesktop)
		}

		if d != nil && d.selection != SelectionFalse {
			d.child = selectedDesktop
			d.env = d.child.env
		} else {
			d = selectedDesktop
		}
	}

	if usrLang != "" {
		conf.Lang = usrLang
	}

	return d
}

// Runs display script, if defined
func runDisplayScript(scriptPath string) {
	if scriptPath != "" {
		if fileIsExecutable(scriptPath) {
			if err := exec.Command(scriptPath).Run(); err != nil {
				logPrint(err)
			}
		} else {
			logPrint(scriptPath + " is not executable.")
		}
	}
}

// Handles keeping information about last login with retry.
func handleLoginRetries(conf *config, usr *sysuser) (result error) {
	// infinite allowed retries, return to avoid writing into file
	if conf.AutologinMaxRetry < 0 {
		return nil
	}

	if conf.Autologin && conf.AutologinSession != "" && conf.AutologinMaxRetry >= 0 {
		retries := 0
		if err := mkDirsForFile(loginRetryFile, 0700); err != nil {
			logPrint(err)
		}

		file, err := os.Open(loginRetryFile)
		if err != nil {
			logPrint(err)
		}
		defer file.Close()

		// Check if last retry was within last 2 seconds
		limit := time.Now().Add(-2 * time.Second)
		if info, err := file.Stat(); err == nil {
			if info.ModTime().After(limit) {
				content, err := os.ReadFile(loginRetryFile)
				if err == nil {
					retries, _ = strconv.Atoi(strings.TrimSpace(string(content)))
				}
				retries++

				if retries >= conf.AutologinMaxRetry {
					result = errors.New("exceeded maximum number of allowed login retries in short period")
					retries = 0
				}
			}
		}
		if err := os.WriteFile(loginRetryFile, []byte(strconv.Itoa(retries)), 0600); err != nil {
			logPrint(err)
		}
	}

	return result
}
