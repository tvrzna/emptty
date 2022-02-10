package src

import (
	"os"
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
	c.XinitrcLaunch = true
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
	c.XinitrcLaunch = true
	_, exec = prepareGuiCommand(u, d, c)
	if strings.Contains(exec, "userHome3") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should not be from homedir: '%s'", exec)
	}

	// Expects no dbus-launch
	c.DbusLaunch = true
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
	c.XinitrcLaunch = false
	d.exec = "/usr/bin/none"
	_, exec = prepareGuiCommand(u, d, c)
	if !strings.HasPrefix(exec, "dbus-launch") {
		t.Errorf("TestPrepareGuiCommandXinitrc: result exec command should start with dbus-launch: '%s'", exec)
	}
}

func TestHandleLoginRetriesInfinite(t *testing.T) {
	c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: -1}
	u := &sysuser{homedir: "/tmp/emptty-test"}

	for i := 0; i < 5; i++ {
		err := handleLoginRetries(c, u)
		if err != nil {
			t.Error("TestHandleLoginRetriesInfinite: No error from handleLoginRetries was expected")
		}
	}

	os.RemoveAll(u.homedir)
}

func TestHandleLoginRetriesNoRetry(t *testing.T) {
	c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: 0}
	u := &sysuser{homedir: "/tmp/emptty-test"}

	for i := 0; i < 5; i++ {
		err := handleLoginRetries(c, u)
		if err != nil {
			break
		}
		if i > 0 {
			t.Error("TestHandleLoginRetriesNoRetry: No retry was expected")
		}
	}

	os.RemoveAll(u.homedir)
}

func TestHandleLoginRetries2Retries(t *testing.T) {
	c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: 2}
	u := &sysuser{homedir: "/tmp/emptty-test"}

	for i := 0; i < 5; i++ {
		err := handleLoginRetries(c, u)
		if err != nil {
			break
		}
		if i > 3 {
			t.Error("TestHandleLoginRetriesNoRetry: No retry was expected")
		}
	}

	os.RemoveAll(u.homedir)
}
