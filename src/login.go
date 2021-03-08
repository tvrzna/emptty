package src

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
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
)

// Login into graphical environment
func login(conf *config) {
	usr := authUser(conf)

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
		conf.lang = usrLang
	}

	defineEnvironment(usr, conf, d)

	runDisplayScript(conf.displayStartScript)

	switch d.env {
	case Wayland:
		wayland(usr, d, conf)
	case Xorg:
		xorg(usr, d, conf)
	}

	closeAuth()

	runDisplayScript(conf.displayStopScript)
}

// Prepares environment and env variables for authorized user.
func defineEnvironment(usr *sysuser, conf *config, d *desktop) {
	defineSpecificEnvVariables(usr)

	usr.setenv(envHome, usr.homedir)
	usr.setenv(envPwd, usr.homedir)
	usr.setenv(envUser, usr.username)
	usr.setenv(envLogname, usr.username)
	usr.setenv(envXdgConfigHome, usr.homedir+"/.config")
	usr.setenv(envXdgRuntimeDir, "/run/user/"+usr.strUid())
	usr.setenv(envXdgSeat, "seat0")
	usr.setenv(envXdgSessionClass, "user")
	usr.setenv(envShell, getUserShell(usr))
	usr.setenv(envLang, conf.lang)
	usr.setenv(envPath, os.Getenv(envPath))

	if d.name != "" {
		usr.setenv(envDesktopSession, d.name)
		usr.setenv(envXdgSessDesktop, d.name)
	} else if d.child != nil && d.child.name != "" {
		usr.setenv(envDesktopSession, d.child.name)
		usr.setenv(envXdgSessDesktop, d.child.name)
	}

	log.Print("Defined Environment")

	// create XDG folder
	if !fileExists(usr.getenv(envXdgRuntimeDir)) {
		err := os.MkdirAll(usr.getenv(envXdgRuntimeDir), 0700)
		handleErr(err)

		// Set owner of XDG folder
		os.Chown(usr.getenv(envXdgRuntimeDir), usr.uid, usr.gid)

		log.Print("Created XDG folder")
	} else {
		log.Print("XDG folder already exists, no need to create")
	}

	os.Chdir(usr.getenv(envPwd))
}

// Reads default shell of authorized user.
func getUserShell(usr *sysuser) string {
	out, err := exec.Command("/usr/bin/getent", "passwd", usr.strUid()).Output()
	handleErr(err)

	ent := strings.Split(strings.TrimSuffix(string(out), "\n"), ":")
	return ent[6]
}

// Prepares and stars Wayland session for authorized user.
func wayland(usr *sysuser, d *desktop, conf *config) {
	// Set environment
	usr.setenv(envXdgSessionType, "wayland")
	log.Print("Defined Wayland environment")

	// start Wayland
	wayland, strExec := prepareGuiCommand(usr, d, conf)
	registerInterruptHandler(wayland)
	log.Print("Starting " + strExec)
	err := wayland.Start()
	handleErr(err)

	// make utmp entry
	utmpEntry := addUtmpEntry(usr.username, wayland.Process.Pid, conf.strTTY(), "")
	log.Print("Added utmp entry")

	wayland.Wait()
	log.Print(strExec + " finished")

	// end utmp entry
	endUtmpEntry(utmpEntry)
	log.Print("Ended utmp entry")
}

// Prepares and starts Xorg session for authorized user.
func xorg(usr *sysuser, d *desktop, conf *config) {
	freeDisplay := strconv.Itoa(getFreeXDisplay())

	// Set environment
	usr.setenv(envXdgSessionType, "x11")
	usr.setenv(envXauthority, usr.getenv(envXdgRuntimeDir)+"/.emptty-xauth")
	usr.setenv(envDisplay, ":"+freeDisplay)
	os.Setenv(envXauthority, usr.getenv(envXauthority))
	os.Setenv(envDisplay, usr.getenv(envDisplay))
	log.Print("Defined Xorg environment")

	// create xauth
	os.Remove(usr.getenv(envXauthority))

	// generate mcookie
	cmd := cmdAsUser(usr, "/usr/bin/mcookie")
	mcookie, err := cmd.Output()
	handleErr(err)
	log.Print("Generated mcookie")

	// generate xauth
	cmd = cmdAsUser(usr, "/usr/bin/xauth", "add", usr.getenv(envDisplay), ".", string(mcookie))
	_, err = cmd.Output()
	handleErr(err)

	log.Print("Generated xauthority")

	// start X
	log.Print("Starting Xorg")

	xorgArgs := []string{"vt" + conf.strTTY(), usr.getenv(envDisplay)}

	if conf.xorgArgs != "" {
		arrXorgArgs := strings.Split(conf.xorgArgs, " ")
		xorgArgs = append(xorgArgs, arrXorgArgs...)
	}

	xorg := exec.Command("/usr/bin/Xorg", xorgArgs...)
	xorg.Env = append(os.Environ())
	xorg.Start()
	if xorg.Process == nil {
		handleStrErr("Xorg is not running")
	}
	log.Print("Started Xorg")

	disp := &xdisplay{}
	disp.dispName = usr.getenv(envDisplay)
	handleErr(disp.openXDisplay())

	// make utmp entry
	utmpEntry := addUtmpEntry(usr.username, xorg.Process.Pid, conf.strTTY(), usr.getenv(envDisplay))
	log.Print("Added utmp entry")

	// start xinit
	xinit, strExec := prepareGuiCommand(usr, d, conf)
	registerInterruptHandler(xinit, xorg)
	log.Print("Starting " + strExec)
	err = xinit.Start()
	if err != nil {
		xorg.Process.Signal(os.Interrupt)
		xorg.Wait()
		handleErr(err)
	}

	xinit.Wait()
	log.Print(strExec + " finished")

	// Stop Xorg
	xorg.Process.Signal(os.Interrupt)
	xorg.Wait()
	log.Print("Interrupted Xorg")

	// Remove auth
	os.Remove(usr.getenv(envXauthority))
	log.Print("Cleaned up xauthority")

	// End utmp entry
	endUtmpEntry(utmpEntry)
	log.Print("Ended utmp entry")
}

// Prepares command for starting GUI.
func prepareGuiCommand(usr *sysuser, d *desktop, conf *config) (*exec.Cmd, string) {
	strExec, allowStartupPrefix := getStrExec(d)

	startScript := false

	if d.selection && d.child != nil {
		strExec = d.path + " " + d.child.exec
	} else {
		if d.env == Xorg && conf.xinitrcLaunch && allowStartupPrefix && !strings.Contains(strExec, ".xinitrc") && fileExists(usr.homedir+"/.xinitrc") {
			startScript = true
			allowStartupPrefix = false
			strExec = usr.homedir + "/.xinitrc " + strExec
		}

		if conf.dbusLaunch && !strings.Contains(strExec, "dbus-launch") && allowStartupPrefix {
			strExec = "dbus-launch " + strExec
		}
	}

	arrExec := strings.Split(strExec, " ")

	var cmd *exec.Cmd
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
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM)
	go handleInterrupt(c, cmds...)
}

// Catch interrupt signal chan and interrupts all mentioned Cmds.
func handleInterrupt(c chan os.Signal, cmds ...*exec.Cmd) {
	<-c
	log.Print("Catched interrupt signal")
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
				log.Print(err)
			}
		} else {
			log.Print(scriptPath + " is not executable.")
		}
	}
}
