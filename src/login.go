package src

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Login into graphical environment
func login(conf *config) {
	interrupted = false
	usr := authUser(conf)

	if err := handleLoginRetries(conf, usr); err != nil {
		closeAuth()
		handleErr(err)
		return
	}

	d, usrLang := loadUserDesktop(usr.homedir)

	if d == nil || (d != nil && d.selection) {
		selectedDesktop, lastDesktop := selectDesktop(usr, conf, d == nil || (d != nil && !d.selection))
		if isLastDesktopForSave(usr, lastDesktop, selectedDesktop) {
			setUserLastSession(usr, selectedDesktop)
		}

		if d != nil && d.selection {
			d.child = selectedDesktop
			d.env = d.child.env
		} else {
			d = selectedDesktop
		}
	}

	if usrLang != "" {
		conf.Lang = usrLang
	}

	runDisplayScript(conf.DisplayStartScript)

	startSession(usr, d, conf)

	closeAuth()

	runDisplayScript(conf.DisplayStopScript)
}

// Runs display script, if defined
func runDisplayScript(scriptPath string) {
	if scriptPath != "" {
		if fileIsExecutable(scriptPath) {
			err := exec.Command(scriptPath).Run()
			if err != nil {
				logPrint(err)
			}
		} else {
			logPrint(scriptPath + " is not executable.")
		}
	}
}

// Handles keeping informations about last login with retry.
func handleLoginRetries(conf *config, usr *sysuser) (result error) {
	// infinite allowed retries, return to avoid writing into file
	if conf.AutologinMaxRetry < 0 {
		return nil
	}

	retries := 0

	if conf.Autologin && conf.AutologinSession != "" && conf.AutologinMaxRetry >= 0 {
		err := mkDirsForFile(usr.getLoginRetryPath(), 0744)
		if err != nil {
			logPrint(err)
		}

		file, err := os.Open(usr.getLoginRetryPath())
		if err != nil {
			logPrint(err)
		}
		defer file.Close()

		// Check if last retry was within last 2 seconds
		limit := time.Now().Add(-2 * time.Second)
		info, err := file.Stat()
		if err == nil {
			if info.ModTime().After(limit) {
				content, err := ioutil.ReadFile(usr.getLoginRetryPath())
				if err == nil {
					retries, _ = strconv.Atoi(strings.TrimSpace(string(content)))
				}
				retries++

				if retries >= conf.AutologinMaxRetry {
					result = errors.New("Exceeded maximum number of allowed login retries in short period.")
				}
			}
		}
	}

	doAsUser(usr, func() {
		err := ioutil.WriteFile(usr.getLoginRetryPath(), []byte(strconv.Itoa(retries)), 0600)
		if err != nil {
			logPrint(err)
		}
	})
	return result
}
