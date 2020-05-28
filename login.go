package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/bgentry/speakeasy"
	"github.com/msteinert/pam"
)

const (
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
func login() {
	usr, trans := authUser()

	var d *desktop
	d, usrLang := loadUserDesktop(usr.homedir)

	if d == nil {
		d = selectDesktop(usr.uid)
	}

	if usrLang != "" {
		conf.lang = usrLang
	}

	defineEnvironment(usr, trans)

	switch d.env {
	case Wayland:
		wayland(usr, d)
	case Xorg:
		xorg(usr, d)
	}

	trans.CloseSession(pam.Silent)
}

// Handle PAM authentication of user.
// If user is successfully authorized, it returns sysuser.
//
// If autologin is enabled, it behaves as user has been authorized.
func authUser() (*sysuser, *pam.Transaction) {
	trans, err := pam.StartFunc("emptty", conf.defaultUser, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			if conf.autologin {
				break
			}
			if conf.defaultUser != "" {
				hostname, _ := os.Hostname()
				fmt.Printf("%s login: %s\n", hostname, conf.defaultUser)
			}
			return speakeasy.Ask("Password: ")
		case pam.PromptEchoOn:
			if conf.autologin {
				break
			}
			hostname, _ := os.Hostname()
			fmt.Printf("%s login: ", hostname)
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return "", err
			}
			return input[:len(input)-1], nil
		case pam.ErrorMsg:
			log.Print(msg)
			return "", nil
		case pam.TextInfo:
			fmt.Println(msg)
			return "", nil
		}
		return "", errors.New("Unrecognized message style")
	})

	err = trans.Authenticate(pam.Silent)
	handleErr(err)
	log.Print("Authenticate OK")

	trans.SetItem(pam.Tty, "tty"+strconv.Itoa(conf.tty))

	trans.OpenSession(pam.Silent)

	pamUsr, _ := trans.GetItem(pam.User)
	usr, _ := user.Lookup(pamUsr)

	return getSysuser(usr), trans
}

// Prepares environment and env variables for authorized user.
// Defines users Uid and Gid for further syscalls.
func defineEnvironment(usr *sysuser, trans *pam.Transaction) {
	envs, _ := trans.GetEnvList()
	for key, value := range envs {
		log.Printf("%s=%s", key, value)
		os.Setenv(key, value)
	}

	os.Setenv(envHome, usr.homedir)
	os.Setenv(envPwd, usr.homedir)
	os.Setenv(envUser, usr.username)
	os.Setenv(envLogname, usr.username)
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

	err = syscall.Setfsuid(usr.uid)
	handleErr(err)
	log.Print("Defined uid")

	err = syscall.Setfsgid(usr.gid)
	handleErr(err)
	log.Print("Defined gid")

	err = syscall.Setgroups(usr.gids)
	handleErr(err)
	log.Print("Defined gids")

	os.Chdir(os.Getenv(envPwd))
}

// Reads default shell of authorized user
func getUserShell(usr *sysuser) string {
	out, err := exec.Command("/usr/bin/getent", "passwd", usr.strUid()).Output()
	handleErr(err)

	ent := strings.Split(strings.TrimSuffix(string(out), "\n"), ":")
	return ent[6]
}

// Prepares and stars Wayland session for authorized user.
func wayland(usr *sysuser, d *desktop) {
	// Set environment
	os.Setenv(envXdgSessionType, "wayland")
	log.Print("Defined Wayland environment")

	// start Wayland
	wayland, strExec := prepareGuiCommand(usr, d)

	log.Print("Starting " + strExec)
	err := wayland.Start()
	handleErr(err)
	wayland.Wait()
	log.Print(strExec + " finished")
}

// Prepares and starts Xorg session for authorized user.
func xorg(usr *sysuser, d *desktop) {
	// Set environment
	os.Setenv(envXdgSessionType, "x11")
	os.Setenv(envXauthority, os.Getenv(envXdgRuntimeDir)+"/.emptty-xauth")
	os.Setenv(envDisplay, ":"+strconv.Itoa(getFreeXDisplay()))
	log.Print("Defined Xorg environment")

	// create xauth
	os.Remove(os.Getenv(envXauthority))
	xauthority, err := os.Create(os.Getenv(envXauthority))
	defer xauthority.Close()
	handleErr(err)
	log.Print("Created xauthority file")

	// generate mcookie
	cmd := exec.Command("/usr/bin/mcookie")
	cmd.Env = append(os.Environ())
	mcookie, err := cmd.Output()
	handleErr(err)
	log.Print("Generated mcookie")

	// create xauth
	cmd = exec.Command("/usr/bin/xauth", "add", os.Getenv(envDisplay), ".", string(mcookie))
	cmd.Env = append(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
	_, err = cmd.Output()
	handleErr(err)

	log.Print("Generated xauthority")

	// start X
	log.Print("Starting Xorg")
	xorg := exec.Command("/usr/bin/Xorg", os.Getenv(envDisplay))
	xorg.Env = append(os.Environ())
	xorg.Start()
	if xorg.Process == nil {
		handleErr(errors.New("Xorg is not running"))
	}
	log.Print("Started Xorg")

	// start xinit
	xinit, strExec := prepareGuiCommand(usr, d)
	log.Print("Starting " + strExec)
	err = xinit.Start()
	if err != nil {
		xorg.Process.Signal(os.Interrupt)
		handleErr(err)
	}
	xinit.Wait()
	log.Print(strExec + " finished")

	// Stop Xorg
	xorg.Process.Signal(os.Interrupt)
	log.Print("Interrupted Xorg")

	// Remove auth
	os.Remove(os.Getenv(envXauthority))
	log.Print("Cleaned up xauthority")
}

// Prepares command for starting GUI
func prepareGuiCommand(usr *sysuser, d *desktop) (*exec.Cmd, string) {
	strExec := getStrExec(d)

	if conf.dbusLaunch && !strings.Contains(strExec, "dbus-launch") {
		strExec = "dbus-launch " + strExec
	}

	arrExec := strings.Split(strExec, " ")

	var cmd *exec.Cmd
	if len(arrExec) > 1 {
		cmd = exec.Command(arrExec[0], arrExec...)
	} else {
		cmd = exec.Command(arrExec[0])
	}

	cmd.Env = append(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: usr.uidu32(), Gid: usr.gidu32(), Groups: usr.gidsu32}
	return cmd, strExec
}

func getStrExec(d *desktop) string {
	if d.exec != "" {
		return d.exec
	}
	return d.path
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
