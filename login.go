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
	envXdgRuntimeDir  = "XDG_RUNTIME_DIR"
	envXdgSessionType = "XDG_SESSION_TYPE"
	envXdgSeat        = "XDG_SEAT"
	envXdgVtnr        = "XDG_VTNR"
	envHome           = "HOME"
	envPwd            = "PWD"
	envUser           = "USER"
	envLogname        = "LOGNAME"
	envXauthority     = "XAUTHORITY"
	envDisplay        = "DISPLAY"
	envShell          = "SHELL"
)

// Login into graphical environment
func login() {
	usr, trans := authUser()
	uid, gid, gids := getUIDandGID(usr)
	defineEnvironment(usr, uid, gid, gids)

	switch conf.environment {
	case Wayland:
		wayland(uint32(uid), uint32(gid), gids)
		break
	case Xorg:
		xorg(uint32(uid), uint32(gid), gids)
	}

	if trans != nil {
		trans.CloseSession(0)
	}
}

// Handle PAM authentication of user.
// If user is successfully authorized, it returns user.User.
//
// If autologin is enabled, it behaves as user has been authorized.
func authUser() (*user.User, *pam.Transaction) {
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

	trans.OpenSession(pam.Silent)

	pamUsr, _ := trans.GetItem(pam.User)
	usr, _ := user.Lookup(pamUsr)

	return usr, trans
}

// Reads Uid and Gid from user.User and returns them as int.
func getUIDandGID(usr *user.User) (int, int, []uint32) {
	uid, _ := strconv.ParseInt(usr.Uid, 10, 32)
	gid, _ := strconv.ParseInt(usr.Gid, 10, 32)
	var gids []uint32
	strGids, err := usr.GroupIds()
	if err == nil {
		for _, val := range strGids {
			value, _ := strconv.ParseInt(val, 10, 32)
			gids = append(gids, uint32(value))
		}
	}
	return int(uid), int(gid), gids
}

// Prepares environment and env variables for authorized user.
// Defines users Uid and Gid for further syscalls.
func defineEnvironment(usr *user.User, uid int, gid int, gids []uint32) {
	os.Setenv(envHome, usr.HomeDir)
	os.Setenv(envPwd, usr.HomeDir)
	os.Setenv(envUser, usr.Username)
	os.Setenv(envLogname, usr.Username)
	os.Setenv(envXdgRuntimeDir, "/run/user/"+usr.Uid)
	os.Setenv(envXdgSeat, "seat0")
	os.Setenv(envShell, getUserShell(usr))

	log.Print("Defined Environment")

	// create XDG folder
	err := os.MkdirAll(os.Getenv(envXdgRuntimeDir), 0700)
	handleErr(err)
	log.Print("Created XDG folder")

	// Set owner of XDG folder
	os.Chown(os.Getenv(envXdgRuntimeDir), uid, gid)

	err = syscall.Setfsuid(uid)
	handleErr(err)
	log.Print("Defined uid")

	err = syscall.Setfsgid(gid)
	handleErr(err)
	log.Print("Defined gid")

	intGids := make([]int, len(gids))
	for _, val := range gids {
		intGids = append(intGids, int(val))
	}

	err = syscall.Setgroups(intGids)
	handleErr(err)

	os.Chdir(os.Getenv(envPwd))
}

// Reads default shell of authorized user
func getUserShell(usr *user.User) string {
	out, err := exec.Command("/bin/getent", "passwd", usr.Uid).Output()
	handleErr(err)

	ent := strings.Split(strings.TrimSuffix(string(out), "\n"), ":")
	return ent[6]
}

// Prepares and stars Wayland session for authorized user.
func wayland(uid uint32, gid uint32, gids []uint32) {
	// Set environment
	os.Setenv(envXdgSessionType, "wayland")
	log.Print("Defined Wayland environment")

	// start Wayland
	log.Print("Starting .winitrc")
	wayland := exec.Command(os.Getenv(envHome) + "/.winitrc")
	wayland.Env = append(os.Environ())
	wayland.SysProcAttr = &syscall.SysProcAttr{}
	wayland.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: gids}
	err := wayland.Start()
	handleErr(err)
	wayland.Wait()
	log.Print(".winitrc finished")
}

// Prepares and starts Xorg session for authorized user.
func xorg(uid uint32, gid uint32, gids []uint32) {
	// Set environment
	os.Setenv(envXdgSessionType, "x11")
	os.Setenv(envXauthority, os.Getenv(envXdgRuntimeDir)+"/.emptty-xauth")
	os.Setenv(envDisplay, ":"+strconv.Itoa(getFreeXDisplay()))
	log.Print("Defined Xorg environment")

	// create xauth
	os.Remove(os.Getenv(envXauthority))
	_, err := os.Create(os.Getenv(envXauthority))
	handleErr(err)
	log.Print("Created xauthority file")

	// generate mcookie
	cmd := exec.Command("/bin/mcookie")
	cmd.Env = append(os.Environ())
	mcookie, err := cmd.Output()
	handleErr(err)
	log.Print("Generated mcookie")

	// create xauth
	cmd = exec.Command("/bin/xauth", "add", os.Getenv(envDisplay), ".", string(mcookie))
	cmd.Env = append(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: gids}
	_, err = cmd.Output()
	handleErr(err)

	log.Print("Generated xauthority")

	// start X
	log.Print("Starting Xorg")
	xorg := exec.Command("/bin/Xorg", os.Getenv(envDisplay))
	xorg.Env = append(os.Environ())
	xorg.Start()
	if xorg.Process == nil {
		handleErr(errors.New("Xorg is not running"))
	}
	log.Print("Started Xorg")

	// start xinit
	log.Print("Starting .xinitrc")
	xinit := exec.Command(os.Getenv(envHome) + "/.xinitrc")
	xinit.Env = append(os.Environ())
	xinit.SysProcAttr = &syscall.SysProcAttr{}
	xinit.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: gids}
	err = xinit.Start()
	if err != nil {
		xorg.Process.Signal(os.Interrupt)
		log.Fatal(err)
	}
	xinit.Wait()
	log.Print(".xinitrc finished")

	// Stop Xorg
	xorg.Process.Signal(os.Interrupt)
	log.Print("Interrupted Xorg")

	// Remove auth
	os.Remove(os.Getenv(envXauthority))
	log.Print("Cleaned up xauthority")
}

// Finds free display for spawning Xorg instance.
func getFreeXDisplay() int {
	for i := 0; i < 32; i++ {
		filename := fmt.Sprintf("/tmp/.X%d-lock", i)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return i
		}
	}
	return 0
}
