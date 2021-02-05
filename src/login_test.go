package src

import (
	"strings"
	"testing"
)

func TestGetStrExec(t *testing.T) {
	d := &desktop{path: "/dev/null", exec: "/usr/bin/none"}

	cmd, isExec := getStrExec(d)
	if !isExec || cmd != "/usr/bin/none" {
		t.Errorf("TestGetStrExec: unexpected result: %s; %t", cmd, isExec)
	}

	d.exec = ""
	cmd, isExec = getStrExec(d)
	if isExec || cmd != "/dev/null" {
		t.Errorf("TestGetStrExec: unexpected result: %s; %t", cmd, isExec)
	}
}

func TestPrepareGuiCommandWithChild(t *testing.T) {
	c := &config{}
	u := &sysuser{uid: 3000, gid: 2000}
	d := &desktop{path: "/dev/null", exec: "/usr/bin/none"}
	d.child = d

	_, exec := prepareGuiCommand(u, d, c)
	if exec != "/usr/bin/none" {
		t.Errorf("TestPrepareGuiCommandWithChild: result exec command is unexpected: '%s'", exec)
	}

	d.selection = true
	_, exec = prepareGuiCommand(u, d, c)
	if exec != "/dev/null /usr/bin/none" {
		t.Errorf("TestPrepareGuiCommandWithChild: result exec command is unexpected: '%s'", exec)
	}
}

func TestPrepareGuiCommandXinitrc(t *testing.T) {
	c := &config{}
	u := &sysuser{uid: 3000, gid: 2000, homedir: getTestingPath("userHome3")}
	d := &desktop{path: "/dev/null", exec: "/usr/bin/none"}

	// No config
	_, exec := prepareGuiCommand(u, d, c)
	if exec != "/usr/bin/none" {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command is unexpected: '%s'", exec)
	}

	// Should be correct
	d.env = Xorg
	c.xinitrcLaunch = true
	_, exec = prepareGuiCommand(u, d, c)
	if !strings.Contains(exec, ".xinitrc") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command does not contain .xinitrc: '%s'", exec)
	}

	// Expects .xinitrc from homedir
	d.env = Wayland
	_, exec = prepareGuiCommand(u, d, c)
	if strings.Contains(exec, u.homedir+".xinitrc") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command contains .xinitrc without homedir: '%s'", exec)
	}

	// Does not expects .xinitrc from homedir
	d.env = Xorg
	d.exec = ""
	c.xinitrcLaunch = true
	_, exec = prepareGuiCommand(u, d, c)
	if strings.Contains(exec, "userHome3") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not be from homedir: '%s'", exec)
	}

	// Expects no dbus-launch
	c.dbusLaunch = true
	d.exec = "/usr/bin/none dbus-launch"
	_, exec = prepareGuiCommand(u, d, c)
	if strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not start with dbus-launch: '%s'", exec)
	}

	//  Expects no dbus-launch
	d.exec = "/usr/bin/none"
	_, exec = prepareGuiCommand(u, d, c)
	if strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not start with dbus-launch: '%s'", exec)
	}

	//  Expects dbus-launch
	c.xinitrcLaunch = false
	d.exec = "/usr/bin/none"
	_, exec = prepareGuiCommand(u, d, c)
	if !strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should start with dbus-launch: '%s'", exec)
	}
}
