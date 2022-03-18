package src

import (
	"os"
	"os/exec"
	"strings"
)

var interrupted bool

// session defines basic functions expected from desktop session
type session interface {
	startCarrier()
	getCarrierPid() int
	finishCarrier() error
}

// commonSession defines structure with data required for starting the session
type commonSession struct {
	session
	usr  *sysuser
	d    *desktop
	conf *config
}

// Starts user's session
func startSession(usr *sysuser, d *desktop, conf *config) {
	s := &commonSession{nil, usr, d, conf}

	switch d.env {
	case Wayland:
		s.session = &waylandSession{s}
	case Xorg:
		s.session = &xorgSession{s, nil}
	}

	s.start()
}

// Performs common start of session
func (s *commonSession) start() {
	s.startCarrier()

	if !s.conf.NoXdgFallback {
		s.usr.setenv(envXdgSessionType, s.d.env.sessionType())
	}

	session, strExec := s.prepareGuiCommand()
	go handleInterrupt(makeInterruptChannel(), session)

	sessionErrLog, sessionErrLogErr := initSessionErrorLogger(s.conf)
	if sessionErrLogErr == nil {
		session.Stderr = sessionErrLog
		defer sessionErrLog.Close()
	} else {
		logPrint(sessionErrLogErr)
	}

	logPrint("Starting " + strExec)
	if err := session.Start(); err != nil {
		s.finishCarrier()
		handleErr(err)
	}

	pid := s.getCarrierPid()
	if pid <= 0 {
		pid = session.Process.Pid
	}

	utmpEntry := addUtmpEntry(s.usr.username, pid, s.conf.strTTY(), s.usr.getenv(envDisplay))
	logPrint("Added utmp entry")

	err := session.Wait()

	carrierErr := s.finishCarrier()

	endUtmpEntry(utmpEntry)
	logPrint("Ended utmp entry")

	if !interrupted && err != nil {
		logPrint(strExec + " finished with error: " + err.Error() + ". For more details see `SESSION_ERROR_LOGGING` in configuration.")
		handleStrErr(s.d.env.string() + " session finished with error, please check logs")
	}

	if !interrupted && carrierErr != nil {
		logPrint(s.d.env.string() + " finished with error: " + carrierErr.Error())
		handleStrErr(s.d.env.string() + " finished with error, please check logs")
	}
}

// Prepares command for starting GUI.
func (s *commonSession) prepareGuiCommand() (cmd *exec.Cmd, strExec string) {
	strExec, allowStartupPrefix := s.d.getStrExec()

	startScript := false

	if s.d.selection && s.d.child != nil {
		strExec = s.d.path + " " + s.d.child.exec
		startScript = true
	} else {
		if s.d.env == Xorg && s.conf.XinitrcLaunch && allowStartupPrefix && !strings.Contains(strExec, ".xinitrc") && fileExists(s.usr.homedir+"/.xinitrc") {
			startScript = true
			allowStartupPrefix = false
			strExec = s.usr.homedir + "/.xinitrc " + strExec
		}

		if s.conf.DbusLaunch && !strings.Contains(strExec, "dbus-launch") && allowStartupPrefix {
			strExec = "dbus-launch " + strExec
		}
	}

	arrExec := strings.Split(strExec, " ")


	if startScript {
		cmd = cmdAsUser(s.usr, s.getLoginShell(), arrExec...)
	} else {
		if len(arrExec) > 1 {
			cmd = cmdAsUser(s.usr, arrExec[0], arrExec...)
		} else {
			cmd = cmdAsUser(s.usr, arrExec[0])
		}
	}

	return cmd, strExec
}

// Gets preffered login shell
func (s *commonSession) getLoginShell() string {
	if s.d.loginShell != "" {
		return s.d.loginShell
	}
	return "/bin/sh"
}

// Catch interrupt signal chan and interrupts Cmd.
func handleInterrupt(c chan os.Signal, cmd *exec.Cmd) {
	<-c
	interrupted = true
	logPrint("Catched interrupt signal")
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}
