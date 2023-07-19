package src

import (
	"os"
	"os/exec"
	"strings"
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
	envXdgCurrDesktop  = "XDG_CURRENT_DESKTOP"
)

// session defines basic functions expected from desktop session
type session interface {
	startCarrier()
	getCarrierPid() int
	finishCarrier() error
}

// commonSession defines structure with data required for starting the session
type commonSession struct {
	session
	auth        authHandle
	d           *desktop
	conf        *config
	dbus        *dbus
	cmd         *exec.Cmd
	interrupted bool
}

// Starts user's session
func createSession(h authHandle, d *desktop, conf *config) *commonSession {
	s := &commonSession{auth: h, d: d, conf: conf}

	switch d.env {
	case Wayland:
		s.session = &waylandSession{s}
	case Xorg:
		s.session = &xorgSession{s, nil}
	}

	return s
}

// Performs common start of session
func (s *commonSession) start() {
	s.defineEnvironment()

	s.startCarrier()

	if !s.conf.NoXdgFallback {
		s.auth.usr().setenv(envXdgSessionType, s.d.env.sessionType())
	}

	if s.conf.AlwaysDbusLaunch {
		s.dbus = &dbus{}
	}

	session, strExec := s.prepareGuiCommand()
	s.cmd = session

	if sessionErrLog, sessionErrLogErr := initSessionErrorLogger(s.conf); sessionErrLogErr == nil {
		session.Stderr = sessionErrLog
		defer sessionErrLog.Close()
	} else {
		logPrint(sessionErrLogErr)
	}

	if s.dbus != nil {
		s.dbus.launch(s.auth.usr())
	}

	logPrint("Starting " + strExec)
	session.Env = s.auth.usr().environ()
	if err := session.Start(); err != nil {
		s.finishCarrier()
		handleErr(err)
	}

	pid := s.getCarrierPid()
	if pid <= 0 {
		pid = session.Process.Pid
	}

	utmpEntry := addUtmpEntry(s.auth.usr().username, pid, s.conf.strTTY(), s.auth.usr().getenv(envDisplay))
	logPrint("Added utmp entry")

	err := session.Wait()

	if s.dbus != nil {
		s.dbus.interrupt()
	}

	carrierErr := s.finishCarrier()

	endUtmpEntry(utmpEntry)
	logPrint("Ended utmp entry")

	if !s.interrupted && err != nil {
		logPrint(strExec + " finished with error: " + err.Error() + ". For more details see `SESSION_ERROR_LOGGING` in configuration.")
		handleStrErr(s.d.env.string() + " session finished with error, please check logs")
	}

	if !s.interrupted && carrierErr != nil {
		logPrint(s.d.env.string() + " finished with error: " + carrierErr.Error())
		handleStrErr(s.d.env.string() + " finished with error, please check logs")
	}
}

// Prepares environment and env variables for authorized user.
func (s *commonSession) defineEnvironment() {
	s.auth.defineSpecificEnvVariables()

	s.auth.usr().setenv(envHome, s.auth.usr().homedir)
	s.auth.usr().setenv(envPwd, s.auth.usr().homedir)
	s.auth.usr().setenv(envUser, s.auth.usr().username)
	s.auth.usr().setenv(envLogname, s.auth.usr().username)
	s.auth.usr().setenv(envUid, s.auth.usr().strUid())
	if !s.conf.NoXdgFallback {
		s.auth.usr().setenvIfEmpty(envXdgConfigHome, s.auth.usr().homedir+"/.config")
		s.auth.usr().setenvIfEmpty(envXdgRuntimeDir, "/run/user/"+s.auth.usr().strUid())
		s.auth.usr().setenvIfEmpty(envXdgSeat, "seat0")
		s.auth.usr().setenv(envXdgSessionClass, "user")
	}
	s.auth.usr().setenv(envShell, s.auth.usr().getShell())
	s.auth.usr().setenvIfEmpty(envLang, s.conf.Lang)
	s.auth.usr().setenvIfEmpty(envPath, os.Getenv(envPath))

	if !s.conf.NoXdgFallback {
		if s.d.name != "" {
			s.auth.usr().setenv(envDesktopSession, s.d.name)
			s.auth.usr().setenv(envXdgSessDesktop, s.d.name)
		} else if s.d.child != nil && s.d.child.name != "" {
			s.auth.usr().setenv(envDesktopSession, s.d.child.name)
			s.auth.usr().setenv(envXdgSessDesktop, s.d.child.name)
		}

		if s.d.desktopNames != "" {
			s.auth.usr().setenv(envXdgCurrDesktop, s.d.desktopNames)
		} else if s.d.child != nil && s.d.child.desktopNames != "" {
			s.auth.usr().setenv(envXdgCurrDesktop, s.d.child.desktopNames)
		}
	}

	logPrint("Defined Environment")

	// create XDG folder
	if !s.conf.NoXdgFallback {
		if !fileExists(s.auth.usr().getenv(envXdgRuntimeDir)) {
			handleErr(os.MkdirAll(s.auth.usr().getenv(envXdgRuntimeDir), 0700))

			// Set owner of XDG folder
			os.Chown(s.auth.usr().getenv(envXdgRuntimeDir), s.auth.usr().uid, s.auth.usr().gid)

			logPrint("Created XDG folder")
		} else {
			logPrint("XDG folder already exists, no need to create")
		}
	}

	os.Chdir(s.auth.usr().getenv(envPwd))
}

// Prepares command for starting GUI.
func (s *commonSession) prepareGuiCommand() (cmd *exec.Cmd, strExec string) {
	strExec, allowStartupPrefix := s.d.getStrExec()

	startScript := s.d.isUser && !allowStartupPrefix

	if allowStartupPrefix && s.conf.XinitrcLaunch && s.d.env == Xorg && !strings.Contains(strExec, ".xinitrc") && fileExists(s.auth.usr().homedir+"/.xinitrc") {
		startScript = true
		strExec = s.auth.usr().homedir + "/.xinitrc " + strExec
	} else if allowStartupPrefix && s.conf.DbusLaunch && !strings.Contains(strExec, "dbus-launch") {
		s.dbus = &dbus{}
	}

	if startScript {
		cmd = cmdAsUser(s.auth.usr(), s.getLoginShell(), strings.Split(strExec, " ")...)
	} else {
		cmd = cmdAsUser(s.auth.usr(), strExec)
	}

	return cmd, strExec
}

// Gets preferred login shell
func (s *commonSession) getLoginShell() string {
	if s.d.loginShell != "" {
		return s.d.loginShell
	}
	return "/bin/sh"
}
