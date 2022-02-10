package src

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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

var interrupted bool

// Login into graphical environment
func login(conf *config) {
	interrupted = false
	usr := authUser(conf)

	if err := handleLoginRetries(conf, usr); err != nil {
		closeAuth()
		handleErr(err)
		return
	}

	var d *desktop
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

// Prepares and stars Wayland session for authorized user.
func wayland(usr *sysuser, d *desktop, conf *config) {
	// Set environment
	if !conf.NoXdgFallback {
		usr.setenv(envXdgSessionType, "wayland")
	}
	logPrint("Defined Wayland environment")

	// start Wayland
	wsession, strExec := prepareGuiCommand(usr, d, conf)
	registerInterruptHandler(wsession)

	sessionErrLog, sessionErrLogErr := initSessionErrorLogger(conf)
	if sessionErrLogErr == nil {
		wsession.Stderr = sessionErrLog
		defer sessionErrLog.Close()
	} else {
		logPrint(sessionErrLogErr)
	}

	logPrint("Starting " + strExec)
	err := wsession.Start()
	handleErr(err)

	// make utmp entry
	utmpEntry := addUtmpEntry(usr.username, wsession.Process.Pid, conf.strTTY(), "")
	logPrint("Added utmp entry")

	logPrint(strExec + " finished")

	// end utmp entry
	endUtmpEntry(utmpEntry)
	logPrint("Ended utmp entry")

	if err = wsession.Wait(); !interrupted && err != nil {
		logPrint(strExec + " finished with error: " + err.Error() + ". For more details see `SESSION_ERROR_LOGGING` in configuration.")
		handleStrErr("Wayland session finished with error, please check logs")
	}
}

// Prepares and starts Xorg session for authorized user.
func xorg(usr *sysuser, d *desktop, conf *config) {
	freeDisplay := strconv.Itoa(getFreeXDisplay())

	// Set environment
	if !conf.NoXdgFallback {
		usr.setenv(envXdgSessionType, "x11")
	}
	if !conf.DefaultXauthority {
		usr.setenv(envXauthority, usr.getenv(envXdgRuntimeDir)+"/.emptty-xauth")
		os.Setenv(envXauthority, usr.getenv(envXauthority))
		os.Remove(usr.getenv(envXauthority))
	}
	usr.setenv(envDisplay, ":"+freeDisplay)
	os.Setenv(envDisplay, usr.getenv(envDisplay))
	logPrint("Defined Xorg environment")

	// generate mcookie
	cmd := cmdAsUser(usr, "/usr/bin/mcookie")
	mcookie, err := cmd.Output()
	handleErr(err)
	logPrint("Generated mcookie")

	// generate xauth
	cmd = cmdAsUser(usr, "/usr/bin/xauth", "add", usr.getenv(envDisplay), ".", string(mcookie))
	_, err = cmd.Output()
	handleErr(err)

	logPrint("Generated xauthority")

	// start X
	logPrint("Starting Xorg")

	var xorgArgs []string
	if conf.RootlessXorg && conf.DaemonMode {
		xorgArgs = []string{"-keeptty", "vt" + conf.strTTY(), usr.getenv(envDisplay)}
	} else {
		xorgArgs = []string{"vt" + conf.strTTY(), usr.getenv(envDisplay)}
	}

	if conf.XorgArgs != "" {
		arrXorgArgs := strings.Split(conf.XorgArgs, " ")
		xorgArgs = append(xorgArgs, arrXorgArgs...)
	}

	var xorg *exec.Cmd
	if conf.RootlessXorg && conf.DaemonMode {
		xorg = cmdAsUser(usr, "/usr/bin/Xorg", xorgArgs...)
		xorg.Env = append(usr.environ())
		err = setTTYOwnership(conf, usr.uid)
		if err != nil {
			logPrint(err)
		}
	} else {
		xorg = exec.Command("/usr/bin/Xorg", xorgArgs...)
		xorg.Env = append(os.Environ())
	}

	xorg.Start()
	if xorg.Process == nil {
		handleStrErr("Xorg is not running")
	}
	logPrint("Started Xorg")

	disp := &xdisplay{}
	disp.dispName = usr.getenv(envDisplay)
	handleErr(disp.openXDisplay())

	// make utmp entry
	utmpEntry := addUtmpEntry(usr.username, xorg.Process.Pid, conf.strTTY(), usr.getenv(envDisplay))
	logPrint("Added utmp entry")

	// start xsession
	xsession, strExec := prepareGuiCommand(usr, d, conf)
	registerInterruptHandler(xsession, xorg)

	sessionErrLog, sessionErrLogErr := initSessionErrorLogger(conf)
	if sessionErrLogErr == nil {
		xsession.Stderr = sessionErrLog
		defer sessionErrLog.Close()
	} else {
		logPrint(sessionErrLogErr)
	}

	logPrint("Starting " + strExec)
	err = xsession.Start()
	if err != nil {
		xorg.Process.Signal(os.Interrupt)
		xorg.Wait()
		handleErr(err)
	}

	errXsession := xsession.Wait()
	logPrint(strExec + " finished")

	// Stop Xorg
	xorg.Process.Signal(os.Interrupt)
	errXorg := xorg.Wait()
	logPrint("Interrupted Xorg")

	// Remove auth
	os.Remove(usr.getenv(envXauthority))
	logPrint("Cleaned up xauthority")

	// End utmp entry
	endUtmpEntry(utmpEntry)
	logPrint("Ended utmp entry")

	if conf.RootlessXorg && conf.DaemonMode {
		err = setTTYOwnership(conf, os.Getuid())
		if err != nil {
			logPrint(err)
		}
	}

	if !interrupted && errXsession != nil {
		logPrint(strExec + " finished with error: " + errXsession.Error() + ". For more details see `SESSION_ERROR_LOGGING` in configuration.")
		handleStrErr("Xorg session finished with error, please check logs")
	}

	if !interrupted && errXorg != nil {
		logPrint("Xorg finished with error: " + errXsession.Error())
		handleStrErr("Xorg finished with error, please check logs")
	}
}

// Sets TTY ownership to defined uid, but keeps the original gid.
func setTTYOwnership(conf *config, uid int) error {
	info, err := os.Stat(conf.ttyPath())
	if err != nil {
		return err
	}
	stat := info.Sys().(*syscall.Stat_t)

	err = os.Chown(conf.ttyPath(), uid, int(stat.Gid))
	if err != nil {
		return err
	}
	err = os.Chmod(conf.ttyPath(), 0620)
	return err
}

// Prepares command for starting GUI.
func prepareGuiCommand(usr *sysuser, d *desktop, conf *config) (cmd *exec.Cmd, strExec string) {
	strExec, allowStartupPrefix := getStrExec(d)

	startScript := false

	if d.selection && d.child != nil {
		strExec = d.path + " " + d.child.exec
		startScript = true
	} else {
		if d.env == Xorg && conf.XinitrcLaunch && allowStartupPrefix && !strings.Contains(strExec, ".xinitrc") && fileExists(usr.homedir+"/.xinitrc") {
			startScript = true
			allowStartupPrefix = false
			strExec = usr.homedir + "/.xinitrc " + strExec
		}

		if conf.DbusLaunch && !strings.Contains(strExec, "dbus-launch") && allowStartupPrefix {
			strExec = "dbus-launch " + strExec
		}
	}

	arrExec := strings.Split(strExec, " ")

	if len(arrExec) > 1 {
		if startScript {
			cmd = cmdAsUser(usr, "/bin/sh", arrExec...)
		} else {
			cmd = cmdAsUser(usr, arrExec[0], arrExec...)
		}
	} else {
		cmd = cmdAsUser(usr, arrExec[0])
	}

	return cmd, strExec
}

// Gets exec path from desktop and returns true, if command allows dbus-launch.
func getStrExec(d *desktop) (string, bool) {
	if d.exec != "" {
		return d.exec, true
	}
	return d.path, false
}

// Finds free display for spawning Xorg instance.
func getFreeXDisplay() int {
	for i := 0; i < 32; i++ {
		filename := fmt.Sprintf("/tmp/.X%d-lock", i)
		if !fileExists(filename) {
			return i
		}
	}
	return 0
}

// Registers interrupt handler, that interrupts all mentioned Cmds.
func registerInterruptHandler(cmds ...*exec.Cmd) {
	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
	go handleInterrupt(c, cmds...)
}

// Catch interrupt signal chan and interrupts all mentioned Cmds.
func handleInterrupt(c chan os.Signal, cmds ...*exec.Cmd) {
	<-c
	interrupted = true
	logPrint("Catched interrupt signal")
	for _, cmd := range cmds {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
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
