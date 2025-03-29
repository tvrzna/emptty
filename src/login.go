package src

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	loginRetryFileBase = "/tmp/emptty/login-retry-"
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
		retriesPath := getLoginRetryPath(conf)
		retries, lastTime := readRetryFile(retriesPath)

		// Check if last retry was within last 2 seconds
		currTime := getUptime()
		limit := currTime - 2
		if lastTime >= limit {
			retries++

			if retries >= conf.AutologinMaxRetry {
				result = errors.New("exceeded maximum number of allowed login retries in short period")
				retries = 0
			}
		}
		writeRetryFile(retriesPath, retries, currTime)
	}

	return result
}

// Parse the retry file at the given path and return time and retry count
func readRetryFile(path string) (retries int, time float64) {
	content, err := os.ReadFile(path)
	if err != nil {
		return retries, time
	}
	contentSlice := strings.Split(string(content), ":")
	contentSliceLen := len(contentSlice)

	if contentSliceLen > 0 && contentSliceLen <= 2 {
		retries, _ = strconv.Atoi(strings.TrimSpace(contentSlice[0]))
		if contentSliceLen == 2 {
			time, _ = strconv.ParseFloat(strings.TrimSpace(contentSlice[1]), 64)
		}
	} else {
		logPrint("Unable to parse the user login retry file")
	}

	return retries, time
}

// Write the given retry count and time to a file at the given path
func writeRetryFile(path string, retries int, time float64) {
	if err := mkDirsForFile(path, 0700); err != nil {
		logPrint(err)
	}

	result := []byte(strconv.Itoa(retries) + ":")
	result = strconv.AppendFloat(result, time, 'f', -1, 64)
	if err := os.WriteFile(path, result, 0600); err != nil {
		logPrint(err)
	}
}

// Attempt to fetch the current device uptime
func getUptime() (uptime float64) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		logPrint("Unable to read /proc/uptime")
		return 0
	}

	slice := strings.Split(string(data), " ")
	uptime, err = strconv.ParseFloat(slice[0], 64)
	if err != nil {
		logPrint("Unable to parse uptime value")
		return 0
	}

	return uptime
}

// Return a tty specific retry file path, future proofing for multi-seat
func getLoginRetryPath(conf *config) string {
	return loginRetryFileBase + conf.strTTY()
}
