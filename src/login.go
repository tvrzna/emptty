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

const (
	envXdgConfigHome   = "XDG_CONFIG_HOME"
	envXdgRuntimeDir   = "XDG_RUNTIME_DIR"
	envXdgSessionId    = "XDG_SESSION_ID"
	envXdgSessionType  = "XDG_SESSION_TYPE"
	envXdgSessionClass = "XDG_SESSION_CLASS"
	envXdgSeat         = "XDG_SEAT"
	envHome            = "HOME"
	envPwd             = "PWD"
	envUser            = "USER"
	envLogname         = "LOGNAME"
	envXauthority      = "XAUTHORITY"
	envDisplay         = "DISPLAY"
	envShell           = "SHELL"
	envLang            = "LANG"
	envPath            = "PATH"
	envDesktopSession  = "DESKTOP_SESSION"
	envXdgSessDesktop  = "XDG_SESSION_DESKTOP"
	envUid             = "UID"
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
		selectedDesktop := selectDesktop(usr, conf)
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

	defineEnvironment(usr, conf, d)

	runDisplayScript(conf.DisplayStartScript)

	switch d.env {
	case Wayland:
		wayland(usr, d, conf)
	case Xorg:
		xorg(usr, d, conf)
	}

	closeAuth()

	runDisplayScript(conf.DisplayStopScript)
}

// Prepares environment and env variables for authorized user.
func defineEnvironment(usr *sysuser, conf *config, d *desktop) {
	defineSpecificEnvVariables(usr)

	usr.setenv(envHome, usr.homedir)
	usr.setenv(envPwd, usr.homedir)
	usr.setenv(envUser, usr.username)
	usr.setenv(envLogname, usr.username)
	usr.setenv(envUid, usr.strUid())
	if !conf.NoXdgFallback {
		usr.setenvIfEmpty(envXdgConfigHome, usr.homedir+"/.config")
		usr.setenvIfEmpty(envXdgRuntimeDir, "/run/user/"+usr.strUid())
		usr.setenvIfEmpty(envXdgSeat, "seat0")
		usr.setenv(envXdgSessionClass, "user")
	}
	usr.setenv(envShell, usr.getShell())
	usr.setenvIfEmpty(envLang, conf.Lang)
	usr.setenvIfEmpty(envPath, os.Getenv(envPath))

	if !conf.NoXdgFallback {
		if d.name != "" {
			usr.setenv(envDesktopSession, d.name)
			usr.setenv(envXdgSessDesktop, d.name)
		} else if d.child != nil && d.child.name != "" {
			usr.setenv(envDesktopSession, d.child.name)
			usr.setenv(envXdgSessDesktop, d.child.name)
		}
	}

	logPrint("Defined Environment")

	// create XDG folder
	if !conf.NoXdgFallback {
		if !fileExists(usr.getenv(envXdgRuntimeDir)) {
			err := os.MkdirAll(usr.getenv(envXdgRuntimeDir), 0700)
			handleErr(err)

			// Set owner of XDG folder
			os.Chown(usr.getenv(envXdgRuntimeDir), usr.uid, usr.gid)

			logPrint("Created XDG folder")
		} else {
			logPrint("XDG folder already exists, no need to create")
		}
	}

	os.Chdir(usr.getenv(envPwd))
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

	if conf.Autologin && conf.AutologinSession != "" && conf.AutologinMaxRetry >= 0 {
		retries := 0
		doAsUser(usr, func() {
			if err := mkDirsForFile(usr.getLoginRetryPath(), 0744); err != nil {
				logPrint(err)
			}
		})

		file, err := os.Open(usr.getLoginRetryPath())
		if err != nil {
			logPrint(err)
		}
		defer file.Close()

		// Check if last retry was within last 2 seconds
		limit := time.Now().Add(-2 * time.Second)
		if info, err := file.Stat(); err == nil {
			if info.ModTime().After(limit) {
				content, err := ioutil.ReadFile(usr.getLoginRetryPath())
				if err == nil {
					retries, _ = strconv.Atoi(strings.TrimSpace(string(content)))
				}
				retries++

				if retries >= conf.AutologinMaxRetry {
					result = errors.New("Exceeded maximum number of allowed login retries in short period.")
					retries = 0
				}
			}
		}
		doAsUser(usr, func() {
			if err := ioutil.WriteFile(usr.getLoginRetryPath(), []byte(strconv.Itoa(retries)), 0600); err != nil {
				logPrint(err)
			}
		})
	}

	return result
}
