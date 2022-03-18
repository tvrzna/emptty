package src

import (
	"strings"
	"testing"
)

func TestPrepareGuiCommandWithChild(t *testing.T) {
	c := &config{}
	u := &sysuser{uid: 3000, gid: 2000}
	d := &desktop{path: "/dev/null", exec: "/usr/bin/none"}
	d.child = d

	s := &commonSession{nil, u, d, c}

	_, exec := s.prepareGuiCommand()
	if exec != "/usr/bin/none" {
		t.Errorf("TestPrepareGuiCommandWithChild: result exec command is unexpected: '%s'", exec)
	}

	d.selection = true
	_, exec = s.prepareGuiCommand()
	if exec != "/dev/null /usr/bin/none" {
		t.Errorf("TestPrepareGuiCommandWithChild: result exec command is unexpected: '%s'", exec)
	}
}

func TestPrepareGuiCommandXinitrc(t *testing.T) {
	c := &config{}
	u := &sysuser{uid: 3000, gid: 2000, homedir: getTestingPath("userHome3")}
	d := &desktop{path: "/dev/null", exec: "/usr/bin/none", loginShell: "/bin/login-shell"}

	s := &commonSession{nil, u, d, c}

	// No config
	_, exec := s.prepareGuiCommand()
	if exec != "/usr/bin/none" {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command is unexpected: '%s'", exec)
	}

	// Should be correct
	d.env = Xorg
	c.XinitrcLaunch = true
	_, exec = s.prepareGuiCommand()
	if !strings.Contains(exec, ".xinitrc") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command does not contain .xinitrc: '%s'", exec)
	}

	// Expects .xinitrc from homedir
	d.env = Wayland
	_, exec = s.prepareGuiCommand()
	if strings.Contains(exec, u.homedir+".xinitrc") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command contains .xinitrc without homedir: '%s'", exec)
	}

	// Does not expects .xinitrc from homedir
	d.env = Xorg
	d.exec = ""
	c.XinitrcLaunch = true
	_, exec = s.prepareGuiCommand()
	if strings.Contains(exec, "userHome3") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not be from homedir: '%s'", exec)
	}

	// Expects no dbus-launch
	c.DbusLaunch = true
	d.exec = "/usr/bin/none dbus-launch"
	cmd, exec := s.prepareGuiCommand()
	if strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not start with dbus-launch: '%s'", exec)
	}
	if !strings.HasPrefix(cmd.String(), d.loginShell) {
		t.Errorf("TestPrepareGuiCommandXinitrc: result cmd command should start with /bin/login-shell: '%s'", cmd.String())
	}

	//  Expects no dbus-launch
	d.exec = "/usr/bin/none"
	d.loginShell = ""
	cmd, exec = s.prepareGuiCommand()
	if strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not start with dbus-launch: '%s'", exec)
	}
	if !strings.HasPrefix(cmd.String(), "/bin/sh") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result cmd command should start with /bin/sh: '%s'", cmd.String())
	}

	//  Expects dbus-launch
	c.XinitrcLaunch = false
	d.exec = "/usr/bin/none"
	_, exec = s.prepareGuiCommand()
	if !strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should start with dbus-launch: '%s'", exec)
	}
}
