package src

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// xorgSession defines structure for xorg
type xorgSession struct {
	*commonSession
	xorg *exec.Cmd
}

// Starts Xorg as carrier for Xorg Session.
func (x *xorgSession) startCarrier() {
	if !x.conf.DefaultXauthority {
		x.usr.setenv(envXauthority, x.usr.getenv(envXdgRuntimeDir)+"/.emptty-xauth")
		os.Remove(x.usr.getenv(envXauthority))
	}

	x.usr.setenv(envDisplay, ":"+x.getFreeXDisplay())

	// generate mcookie
	cmd := cmdAsUser(x.usr, "/usr/bin/mcookie")
	mcookie, err := cmd.Output()
	handleErr(err)
	logPrint("Generated mcookie")

	// generate xauth
	cmd = cmdAsUser(x.usr, "/usr/bin/xauth", "add", x.usr.getenv(envDisplay), ".", string(mcookie))
	_, err = cmd.Output()
	handleErr(err)
	logPrint("Generated xauthority")

	// start X
	logPrint("Starting Xorg")

	var xorgArgs []string
	if x.conf.RootlessXorg && x.conf.DaemonMode {
		xorgArgs = []string{"-keeptty", "vt" + x.conf.strTTY(), x.usr.getenv(envDisplay)}
	} else {
		xorgArgs = []string{"vt" + x.conf.strTTY(), x.usr.getenv(envDisplay)}
	}

	if x.conf.XorgArgs != "" {
		arrXorgArgs := strings.Split(x.conf.XorgArgs, " ")
		xorgArgs = append(xorgArgs, arrXorgArgs...)
	}

	if x.conf.RootlessXorg && x.conf.DaemonMode {
		x.xorg = cmdAsUser(x.usr, "/usr/bin/Xorg", xorgArgs...)
		x.xorg.Env = append(x.usr.environ())
		if err := x.setTTYOwnership(x.conf, x.usr.uid); err != nil {
			logPrint(err)
		}
	} else {
		x.xorg = exec.Command("/usr/bin/Xorg", xorgArgs...)
		os.Setenv(envDisplay, x.usr.getenv(envDisplay))
		os.Setenv(envXauthority, x.usr.getenv(envXauthority))
		x.xorg.Env = append(os.Environ())
	}

	x.xorg.Start()
	if x.xorg.Process == nil {
		handleStrErr("Xorg is not running")
	}
	logPrint("Started Xorg")

	handleErr(openXDisplay(x.usr.getenv(envDisplay)))
}

// Gets Xorg Pid as int
func (x *xorgSession) getCarrierPid() int {
	if x.xorg == nil {
		handleStrErr("Xorg is not running")
	}
	return x.xorg.Process.Pid
}

// Finishes Xorg as carrier for Xorg Session
func (x *xorgSession) finishCarrier() error {
	// Stop Xorg
	x.xorg.Process.Signal(os.Interrupt)
	err := x.xorg.Wait()
	logPrint("Interrupted Xorg")

	// Remove auth
	os.Remove(x.usr.getenv(envXauthority))
	logPrint("Cleaned up xauthority")

	// Revert rootless TTY ownership
	if x.conf.RootlessXorg && x.conf.DaemonMode {
		if err := x.setTTYOwnership(x.conf, os.Getuid()); err != nil {
			logPrint(err)
		}
	}

	return err
}

// Sets TTY ownership to defined uid, but keeps the original gid.
func (x *xorgSession) setTTYOwnership(conf *config, uid int) error {
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

// Finds free display for spawning Xorg instance.
func (x *xorgSession) getFreeXDisplay() string {
	for i := 0; i < 32; i++ {
		filename := fmt.Sprintf("/tmp/.X%d-lock", i)
		if !fileExists(filename) {
			return strconv.Itoa(i)
		}
	}
	return "0"
}
