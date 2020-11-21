package main

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
)

// Login into graphical environment
func login(conf *config) {
	usr := authUser(conf)

	var d *desktop
	d, usrLang := loadUserDesktop(usr.homedir)

	if d == nil {
		d = selectDesktop(usr, conf)
	}

	if usrLang != "" {
		conf.lang = usrLang
	}

	defineEnvironment(usr, conf)

	switch d.env {
	case Wayland:
		wayland(usr, d, conf)
	case Xorg:
		xorg(usr, d, conf)
	}

	closeAuth()
}

// Prepares environment and env variables for authorized user.
// Defines users Uid and Gid for further syscalls.
func defineEnvironment(usr *sysuser, conf *config) {
	defineSpecificEnvVariables()

	os.Setenv(envHome, usr.homedir)
	os.Setenv(envPwd, usr.homedir)
	os.Setenv(envUser, usr.username)
	os.Setenv(envLogname, usr.username)
	os.Setenv(envXdgConfigHome, usr.homedir+"/.config")
	os.Setenv(envXdgRuntimeDir, "/run/user/"+usr.strUid())
	os.Setenv(envXdgSeat, "seat0")
	os.Setenv(envXdgSessionClass, "user")
	os.Setenv(envShell, getUserShell(usr))
	os.Setenv(envLang, conf.lang)

	log.Print("Defined Environment")

	// create XDG folder
	err := os.MkdirAll(os.Getenv(envXdgRuntimeDir), 0700)
	handleErr(err)
	log.Print("Created XDG folder")

	// Set owner of XDG folder
	os.Chown(os.Getenv(envXdgRuntimeDir), usr.uid, usr.gid)

	os.Chdir(os.Getenv(envPwd))
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
	os.Setenv(envXdgSessionType, "wayland")
	log.Print("Defined Wayland environment")

	// start Wayland
	wayland, strExec := prepareGuiCommand(usr, d, conf)
	registerInterruptHandler(nil, wayland)

	log.Print("Starting " + strExec)
	err := wayland.Start()
	handleErr(err)

	// make utmp entry
	utmpEntry := addUtmpEntry(usr.username, wayland.Process.Pid, conf.strTTY())
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
	os.Setenv(envXdgSessionType, "x11")
	os.Setenv(envXauthority, os.Getenv(envXdgRuntimeDir)+"/.emptty-xauth")
	os.Setenv(envDisplay, ":"+freeDisplay)
	log.Print("Defined Xorg environment")

	// create xauth
	os.Remove(os.Getenv(envXauthority))
	xauthority, err := os.Create(os.Getenv(envXauthority))
	os.Chown(os.Getenv(envXauthority), usr.uid, usr.gid)
	xauthority.Close()
	handleErr(err)
	log.Print("Created xauthority file")

	// generate mcookie
	cmd := exec.Command("/usr/bin/mcookie")
	cmd.Env = append(os.Environ())
	mcookie, err := cmd.Output()
	handleErr(err)
	log.Print("Generated mcookie")

	// generate xauth
	cmd = exec.Command("/usr/bin/xauth", "add", os.Getenv(envDisplay), ".", string(mcookie))
	cmd.Env = append(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
	_, err = cmd.Output()
	handleErr(err)

	log.Print("Generated xauthority")

	// start X
	log.Print("Starting Xorg")

	xorgArgs := []string{"vt" + conf.strTTY(), os.Getenv(envDisplay)}

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
	handleErr(disp.openXDisplay())
	defer disp.closeXDisplay()

	// make utmp entry
	utmpEntry := addUtmpEntry(usr.username, xorg.Process.Pid, conf.strTTY())
	log.Print("Added utmp entry")

	// start xinit
	xinit, strExec := prepareGuiCommand(usr, d, conf)
	registerInterruptHandler(disp, xorg, xinit)
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
	os.Remove(os.Getenv(envXauthority))
	log.Print("Cleaned up xauthority")

	// End utmp entry
	endUtmpEntry(utmpEntry)
	log.Print("Ended utmp entry")
}

// Prepares command for starting GUI.
func prepareGuiCommand(usr *sysuser, d *desktop, conf *config) (*exec.Cmd, string) {
	strExec, allowStartupPrefix := getStrExec(d)

	startScript := false

	if d.env == Xorg && conf.xinitrcLaunch && allowStartupPrefix && !strings.Contains(strExec, ".xinitrc") && fileExists(usr.homedir+"/.xinitrc") {
		startScript = true
		allowStartupPrefix = false
		strExec = usr.homedir + "/.xinitrc " + strExec
	}

	if conf.dbusLaunch && !strings.Contains(strExec, "dbus-launch") && allowStartupPrefix {
		strExec = "dbus-launch " + strExec
	}

	arrExec := strings.Split(strExec, " ")

	var cmd *exec.Cmd
	if len(arrExec) > 1 {
		if startScript {
			cmd = exec.Command("/bin/sh", arrExec...)
		} else {
			cmd = exec.Command(arrExec[0], arrExec...)
		}
	} else {
		cmd = exec.Command(arrExec[0])
	}

	cmd.Env = append(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
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
func registerInterruptHandler(disp *xdisplay, cmds ...*exec.Cmd) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go handleInterrupt(c, disp, cmds...)
}

// Catch interrupt signal chan and interrupts all mentioned Cmds.
func handleInterrupt(c chan os.Signal, disp *xdisplay, cmds ...*exec.Cmd) {
	<-c
	log.Print("Catched interrupt signal")
	for _, cmd := range cmds {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
	if disp != nil {
		disp.closeXDisplay()
	}
	os.Exit(1)
}
