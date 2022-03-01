package src

import (
	"fmt"
	"os"
	"os/user"
	"testing"
)

func TestStringifyEnv(t *testing.T) {
	if Xorg.stringify() != constEnvXorg {
		t.Error("TestStringifyEnv: wrong value for Xorg env")
	}

	if Wayland.stringify() != constEnvWayland {
		t.Error("TestStringifyEnv: wrong value for Wayland env")
	}

	if Custom.stringify() != constEnvXorg {
		t.Error("TestStringifyEnv: wrong value for Custom env")
	}
}

func TestStringEnv(t *testing.T) {
	if Undefined.string() != constEnvSUndefined {
		t.Error("TestStringEnv: wrong value for Xorg env")
	}

	if Xorg.string() != constEnvSXorg {
		t.Error("TestStringEnv: wrong value for Xorg env")
	}

	if Wayland.string() != constEnvSWayland {
		t.Error("TestStringEnv: wrong value for Xorg env")
	}

	if Custom.string() != constEnvSCustom {
		t.Error("TestStringEnv: wrong value for Xorg env")
	}

	if UserCustom.string() != constEnvSUserCustom {
		t.Error("TestStringEnv: wrong value for Xorg env")
	}
}

func TestPrintDesktops(t *testing.T) {
	desktops := []*desktop{{name: "a", envOrigin: Xorg},
		{name: "b", envOrigin: Wayland},
		{name: "c", envOrigin: Custom},
		{name: "d", envOrigin: UserCustom}}

	var result string
	conf := &config{}

	conf.VerticalSelection = false
	conf.IdentifyEnvs = false
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "[0] a, [1] b, [2] c, [3] d" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=false, IdentifyEnvs=false")
	}

	conf.VerticalSelection = true
	conf.IdentifyEnvs = false
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "[0] a\n[1] b\n[2] c\n[3] d" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=true, IdentifyEnvs=false")
	}

	conf.VerticalSelection = false
	conf.IdentifyEnvs = true
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "|Xorg| [0] a  |Wayland| [1] b  |Custom| [2] c  |User Custom| [3] d" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=false, IdentifyEnvs=true")
	}

	conf.VerticalSelection = true
	conf.IdentifyEnvs = true
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "|Xorg|\n[0] a\n\n|Wayland|\n[1] b\n\n|Custom|\n[2] c\n\n|User Custom|\n[3] d" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=true, IdentifyEnvs=true")
	}

}

func TestParseEnv(t *testing.T) {
	var env enEnvironment

	env = parseEnv("", "xorg")
	if env != Xorg {
		t.Error("TestParseEnv: wrong default value")
	}

	env = parseEnv("xorg", "wayland")
	if env != Xorg {
		t.Error("TestParseEnv: wrong parsed value for wayland")
	}

	env = parseEnv("wayland", "xorg")
	if env != Wayland {
		t.Error("TestParseEnv: wrong parsed value for wayland")
	}

	env = parseEnv("aaa", "bbb")
	if env != Xorg {
		t.Error("TestParseEnv: wrong fallback value")
	}
}

func TestLoadUserDesktop(t *testing.T) {
	loadUserDesktop(getTestingPath("userHome2"))
	d, _ := loadUserDesktop(getTestingPath("userHome"))

	fmt.Println(d.exec)

	if d.exec != "none" {
		t.Error("TestLoadUserDesktop: wrong EXEC value")
	}

	if d.selection {
		t.Error("TestLoadUserDesktop: wrong SELECTION value")
	}

	if d.env != Xorg {
		t.Error("TestLoadUserDesktop: wrong ENVIRONMENT value")
	}

	if d.name != "window-manager" {
		t.Error("TestLoadUserDesktop: wrong NAME value")
	}

	if !d.isUser {
		t.Error("TestLoadUserDesktop: wrong isUser value")
	}

	readOutput(func() {
		d, _ = loadUserDesktop(getTestingPath("userHome3"))
		if d == nil || d.exec != "" || d.name != "" {
			t.Error("TestLoadUserDesktop: No desktop returned, selection does not return empty desktop or exec/name are not empty.")
		}
	})

	d, _ = loadUserDesktop("/dev/null")
	if d != nil {
		t.Error("TestLoadUserDesktop: No desktop should be returned, no data available")
	}
}

func TestGetDesktop(t *testing.T) {
	d := getDesktop(getTestingPath("userHome/.config/emptty"), Custom)

	if d.exec != "none" {
		t.Error("TestLoadUserDesktop: wrong EXEC value")
	}

	if d.selection {
		t.Error("TestLoadUserDesktop: wrong SELECTION value")
	}

	if d.env != Xorg {
		t.Error("TestLoadUserDesktop: wrong ENVIRONMENT value")
	}

	if d.name != "window-manager" {
		t.Error("TestLoadUserDesktop: wrong NAME value")
	}

	if d.isUser {
		t.Error("TestLoadUserDesktop: wrong isUser value")
	}
}

func TestGetUserLastSession(t *testing.T) {
	usr := &sysuser{}
	usr.homedir = getTestingPath("userHome2")
	getUserLastSession(usr)

	usr.homedir = getTestingPath("userHome")
	s := getUserLastSession(usr)

	if s.env != Wayland {
		t.Error("TestGetUserLastSession: wrong env value")
	}

	if s.exec != "/usr/bin/none" {
		t.Error("TestGetUserLastSession: wrong exec value")
	}
}

func TestGetLastDesktop(t *testing.T) {
	usr := &sysuser{}
	usr.homedir = getTestingPath("userHome")

	desktops := []*desktop{{exec: "/usr/bin/none", env: Xorg}, {exec: "/usr/bin/none", env: Wayland}, {exec: "/usr/bin/none2", env: Wayland}}

	if getLastDesktop(usr, desktops) != 1 {
		t.Error("TestGetLastDesktop: expected different index")
	}
}

func TestListDesktops(t *testing.T) {
	desktops := listDesktops(getTestingPath("userHome"), Custom)
	if len(desktops) > 0 {
		t.Error("TestListDesktops: no desktop was expected")
	}

	desktops = listDesktops(getTestingPath("desktops"), Custom)
	if len(desktops) == 0 {
		t.Error("TestListDesktops: desktops were expected")
	}

	for _, d := range desktops {
		if d.name == "Desktop1" && d.exec != "/usr/bin/desktop1" {
			t.Error("TestListDesktops: wrongly loaded desktop")
		}
	}
}

func TestIsLastDesktopForSave(t *testing.T) {
	currentDesktop := &desktop{exec: "/usr/bin/none", env: Wayland}
	lastDesktop := &desktop{exec: "/usr/bin/none", env: Wayland}

	usr := &sysuser{}
	usr.homedir = "/dev/null"

	if !isLastDesktopForSave(usr, lastDesktop, currentDesktop) {
		t.Error("TestIsLastDesktopForSave: file not exists and doesn't need to save")
	}

	usr.homedir = getTestingPath("userHome")

	if isLastDesktopForSave(usr, lastDesktop, currentDesktop) {
		t.Error("TestIsLastDesktopForSave: desktops should not need to save")
	}

	lastDesktop.env = Xorg
	if !isLastDesktopForSave(usr, lastDesktop, currentDesktop) {
		t.Error("TestIsLastDesktopForSave: desktop should be saved, env is different")
	}
}

func TestSetUserLastSession(t *testing.T) {
	d := &desktop{exec: "/usr/bin/none", env: Wayland}

	currentUser, _ := user.Current()
	usr := getSysuser(currentUser)
	usr.homedir = "/tmp/emptty-test/"

	setUserLastSession(usr, d)

	if !fileExists(usr.homedir + pathLastSession) {
		t.Error("TestSetUserLastSession: last session is not being saved")
	}

	os.RemoveAll(usr.homedir)
}

func TestListAllDesktops(t *testing.T) {
	usr := &sysuser{}
	usr.homedir = getTestingPath("userHome2")

	desktops := listAllDesktops(usr, getTestingPath("desktops"), getTestingPath("desktops"))

	if len(desktops) != 6 {
		t.Error("TestListAllDesktops: unexpected count of desktops, 6 expected")
	}

	usr.homedir = "/dev/null"

	desktops = listAllDesktops(usr, "/dev/null", "/dev/null")
	if len(desktops) != 0 {
		t.Error("TestListAllDesktops: unexpected count of desktops, 0 expected")
	}
}

func TestFindAutoselectDesktop(t *testing.T) {
	usr := &sysuser{}
	usr.homedir = getTestingPath("userHome2")

	desktops := listAllDesktops(usr, getTestingPath("desktops"), getTestingPath("desktops"))

	d1 := findAutoselectDesktop("CustomDesktop1", desktops)
	if d1 == nil || d1.name != "CustomDesktop1" {
		t.Error("TestFindAutoselectDesktop: could not find desktop by its name")
	}

	d2 := findAutoselectDesktop("custom-desktop2", desktops)
	if d2 == nil || d2.name != "CustomDesktop2" {
		t.Error("TestFindAutoselectDesktop: could not find desktop by its exec")
	}

	d3 := findAutoselectDesktop("UnknownDesktop", desktops)
	if d3 != nil {
		t.Error("TestFindAutoselectDesktop: found desktop, that should be uknown")
	}
}

func TestGetStrExec(t *testing.T) {
	d := &desktop{path: "/dev/null", exec: "/usr/bin/none"}

	cmd, isExec := d.getStrExec()
	if !isExec || cmd != "/usr/bin/none" {
		t.Errorf("TestGetStrExec: unexpected result: %s; %t", cmd, isExec)
	}

	d.exec = ""
	cmd, isExec = d.getStrExec()
	if isExec || cmd != "/dev/null" {
		t.Errorf("TestGetStrExec: unexpected result: %s; %t", cmd, isExec)
	}
}
