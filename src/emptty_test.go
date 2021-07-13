package src

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	buildVersion = ""

	if !strings.HasPrefix(getVersion(), version) {
		t.Error("TestGetVersion: version does not start with constant")
	}

	buildVersion = "testing-version"
	if !strings.HasPrefix(getVersion(), buildVersion[1:]) {
		t.Error("TestGetVersion: version does not start with defined version")
	}
}

func TestPrintHelp(t *testing.T) {
	output := readOutput(func() {
		printHelp()
	})

	if len(output) == 0 {
		t.Error("TestPrintHelp: help does not return text")
	}
}

func TestProcessArgs(t *testing.T) {
	conf1 := &config{}
	processArgs([]string{"-d"}, conf1)
	if !conf1.daemonMode {
		t.Error("TestProcessArgs: daemon mode was expected")
	}

	conf2 := &config{tty: 77}
	processArgs([]string{"-t"}, conf2)
	if conf2.daemonMode {
		t.Error("TestProcessArgs: daemon mode was not expected")
	}
	if conf2.tty != 77 {
		t.Error("TestProcessArgs: tty number should not be touched")
	}

	processArgs([]string{"-t", "2"}, conf2)
	if conf2.tty != 2 {
		t.Errorf("TestProcessArgs: expected tty number was 2, but was %d", conf2.tty)
	}

	conf3 := &config{}
	processArgs([]string{"-u"}, conf3)
	if conf3.defaultUser != "" {
		t.Error("TestProcessArgs: no default user was expected")
	}

	processArgs([]string{"-u", "emptty"}, conf3)
	if conf3.defaultUser != "emptty" {
		t.Errorf("TestProcessArgs: expected default user was 'emptty', but was '%s'", conf3.defaultUser)
	}

	conf4 := &config{}
	processArgs([]string{}, conf4)
	if conf4.autologin || conf4.autologinSession != "" {
		t.Error("TestProcessArgs: unexpected value for autologin or autologinSession")
	}

	processArgs([]string{"-a"}, conf4)
	if !conf4.autologin || conf4.autologinSession != "" {
		t.Error("TestProcessArgs: unexpected value for autologin or autologinSession")
	}

	conf4.autologin = false
	processArgs([]string{"-a", "-t", "7"}, conf4)
	if !conf4.autologin || conf4.autologinSession != "" {
		t.Error("TestProcessArgs: unexpected value for autologin or autologinSession")
	}

	conf4.autologin = false
	processArgs([]string{"--autologin", "sway"}, conf4)
	if !conf4.autologin || conf4.autologinSession != "sway" {
		t.Errorf("TestProcessArgs: unexpected value for autologin (is '%t') or autologinSession (is '%s')", conf4.autologin, conf4.autologinSession)
	}

}

func TestNextArg(t *testing.T) {
	args := []string{"one", "two", "three", "four"}

	nextArg(args, 0, func(val string) {
		if val != "two" {
			t.Error("TestNextArg: unexpected next argument")
		}
	})

	nextArg(args, 0, nil)

	nextArg(args, 5, func(val string) {
		t.Error("TestNextArg: index out of bound")
	})

	nextArg(args, 3, func(val string) {
		t.Error("TestNextArg: unexpected next argument")
	})
}
