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
		{name: "d", envOrigin: UserCustom},
		{name: "e", envOrigin: Xorg},
		{name: "f", envOrigin: Wayland},
		{name: "g", envOrigin: Custom},
		{name: "h", envOrigin: UserCustom},
		{name: "i", envOrigin: Xorg},
		{name: "j", envOrigin: Wayland},
		{name: "k", envOrigin: Custom}}

	var result string
	conf := &config{}

	conf.VerticalSelection = false
	conf.IdentifyEnvs = false
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "[0] a, [1] b, [2] c, [3] d, [4] e, [5] f, [6] g, [7] h, [8] i, [9] j, [10] k" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=false, IdentifyEnvs=false")
	}

	conf.VerticalSelection = true
	conf.IdentifyEnvs = false
	conf.IndentSelection = 0
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "[0] a\n[1] b\n[2] c\n[3] d\n[4] e\n[5] f\n[6] g\n[7] h\n[8] i\n[9] j\n[10] k" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=true, IdentifyEnvs=false, IndentSelection=0")
	}

	conf.VerticalSelection = false
	conf.IdentifyEnvs = true
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "|Xorg| [0] a  |Wayland| [1] b  |Custom| [2] c  |User Custom| [3] d  |Xorg| [4] e  |Wayland| [5] f  |Custom| [6] g  |User Custom| [7] h  |Xorg| [8] i  |Wayland| [9] j  |Custom| [10] k" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=false, IdentifyEnvs=true")
	}

	conf.VerticalSelection = true
	conf.IdentifyEnvs = true
	conf.IndentSelection = 0
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "|Xorg|\n[0] a\n\n|Wayland|\n[1] b\n\n|Custom|\n[2] c\n\n|User Custom|\n[3] d\n\n|Xorg|\n[4] e\n\n|Wayland|\n[5] f\n\n|Custom|\n[6] g\n\n|User Custom|\n[7] h\n\n|Xorg|\n[8] i\n\n|Wayland|\n[9] j\n\n|Custom|\n[10] k" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=true, IdentifyEnvs=true, IndentSelection=0")
	}

	conf.VerticalSelection = true
	conf.IdentifyEnvs = false
	conf.IndentSelection = 4
	result = readOutput(func() {
		printDesktops(conf, desktops)
	})
	if result != "     [0] a\n     [1] b\n     [2] c\n     [3] d\n     [4] e\n     [5] f\n     [6] g\n     [7] h\n     [8] i\n     [9] j\n    [10] k" {
		t.Error("TestPrintDesktops: wrong output for VerticalSelection=true, IndentSelection=4")
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

	if d.selection != SelectionFalse {
		t.Error("TestLoadUserDesktop: wrong SELECTION value")
	}

	if d.env != Wayland {
		t.Error("TestLoadUserDesktop: wrong ENVIRONMENT value")
	}

	if d.name != "window-manager" {
		t.Error("TestLoadUserDesktop: wrong NAME value")
	}

	if !d.isUser {
		t.Error("TestLoadUserDesktop: wrong isUser value")
	}

	if d.loginShell == "" {
		t.Error("TestLoadUserDesktop: wrong loginShell value")
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
	d := getDesktop(getTestingPath("desktops/desktop2.desktop"), Custom)

	if d.exec != "/usr/bin/desktop2" {
		t.Error("TestLoadUserDesktop: wrong EXEC value")
	}

	if d.selection != SelectionFalse {
		t.Error("TestLoadUserDesktop: wrong SELECTION value")
	}

	if d.env != Wayland {
		t.Error("TestLoadUserDesktop: wrong ENVIRONMENT value")
	}

	if d.name != "Desktop2" {
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
	desktops := listDesktops(Custom, getTestingPath("userHome"))
	if len(desktops) > 0 {
		t.Error("TestListDesktops: no desktop was expected")
	}

	desktops = listDesktops(Custom, getTestingPath("desktops"))
	if len(desktops) == 0 {
		t.Error("TestListDesktops: desktops were expected")
	}

	for _, d := range desktops {
		if d.name == "Desktop1" && (d.exec != "/usr/bin/desktop1" || d.desktopNames != "Desk1:DESKTOP_1") {
			t.Error("TestListDesktops: wrongly loaded desktop")
		}
		if d.name == "Desktop2" && d.env != Wayland {
			t.Error("TestListDesktops: wrongly loaded desktop, environment is not parsed correctly")
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

	d1 := findAutoselectDesktop("CustomDesktop1", Undefined, desktops)
	if d1 == nil || d1.name != "CustomDesktop1" {
		t.Error("TestFindAutoselectDesktop: could not find desktop by its name")
	}

	d2 := findAutoselectDesktop("custom-desktop2", Xorg, desktops)
	if d2 == nil || d2.name != "CustomDesktop2" {
		t.Error("TestFindAutoselectDesktop: could not find desktop by its exec")
	}

	d3 := findAutoselectDesktop("unknowndESktop", Undefined, desktops)
	if d3 != nil {
		t.Error("TestFindAutoselectDesktop: found desktop, that should be unknown")
	}

	d4 := findAutoselectDesktop("UnknownDesktop", Wayland, desktops)
	if d4 != nil {
		t.Error("TestFindAutoselectDesktop: found desktop, that should be unknown")
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

	d = &desktop{path: "/dev/null", exec: "/usr/bin/none", selection: SelectionAuto, child: d}
	cmd, isExec = d.getStrExec()
	if isExec || cmd != d.path+" "+d.child.exec {
		t.Errorf("TestGetStrExec: unexpected result: %s; %t", cmd, isExec)
	}
}

func TestGetDesktopBaseExec(t *testing.T) {
	exec1, args1 := getDesktopBaseExec("/usr/bin/shell ")
	if exec1 != "shell" || args1 != "" {
		t.Error("TestGetDesktopBaseExec: wrong value for exec1")
	}

	exec2, args2 := getDesktopBaseExec("/usr/bin/shell --argument1 --argument2='none' -a3 some")
	if exec2 != "shell" || args2 != "--argument1 --argument2='none' -a3 some" {
		t.Error("TestGetDesktopBaseExec: wrong value for exec2")
	}

	exec3, args3 := getDesktopBaseExec("shell --argument1 --argument2='none' -a3 some")
	if exec3 != "shell" || args3 != "--argument1 --argument2='none' -a3 some" {
		t.Error("TestGetDesktopBaseExec: wrong value for exec3")
	}

	exec4, args4 := getDesktopBaseExec("shell")
	if exec4 != "shell" || args4 != "" {
		t.Error("TestGetDesktopBaseExec: wrong value for exec4")
	}

	exec5, args5 := getDesktopBaseExec(" / ")
	if exec5 != "" || args5 != "" {
		t.Error("TestGetDesktopBaseExec: wrong value for exec5")
	}
}

func TestGetDesktopName(t *testing.T) {
	d := getDesktop(getTestingPath("desktops/desktop1.desktop"), Custom)

	if desktopName := d.getDesktopName(); desktopName != "Desk1" {
		t.Errorf("TestGetDesktopName: desktop1 got unexpected desktop name '%s'", d.getDesktopName())
	}

	d = getDesktop(getTestingPath("desktops/desktop2.desktop"), Custom)

	if desktopName := d.getDesktopName(); desktopName != "Desktop2" {
		t.Errorf("TestGetDesktopName: desktop2 got unexpected desktop name '%s'", d.getDesktopName())
	}
}
