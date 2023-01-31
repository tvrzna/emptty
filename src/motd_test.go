package src

import (
	"os"
	"strings"
	"testing"
)

func TestPrintDefaultMotd(t *testing.T) {
	output := readOutput(func() {
		printDefaultMotd()
	})

	if output != defaultMotd+"\n\n" {
		t.Error("TestPrintDefaultMotd: default motd does not match")
	}
}

func TestMotdDynamicNotEnabled(t *testing.T) {
	c := &config{PrintMotd: true, DynamicMotd: true, DynamicMotdPath: getTestingPath("motd-dynamic.sh"), MotdPath: getTestingPath("motd-static")}

	output := readOutput(func() {
		printMotd(c)
	})

	if !strings.HasPrefix(output, "This is static motd") {
		t.Error("TestMotdDynamicNotEnabled: unexpected result")
	}

	f, _ := os.Stat(getTestingPath("motd-dynamic.sh"))
	originalMode := f.Mode()
	defer os.Chmod(getTestingPath("motd-dynamic.sh"), originalMode)

	os.Chmod(getTestingPath("motd-dynamic.sh"), 0755)
	c.DynamicMotd = false
	c.MotdPath = ""

	output = readOutput(func() {
		printMotd(c)
	})

	if !strings.HasPrefix(output, defaultMotd) {
		t.Error("TestMotdDynamicNotEnabled: unexpected result")
	}
}

func TestMotdDynamic(t *testing.T) {
	c := &config{PrintMotd: true, DynamicMotd: true, DynamicMotdPath: getTestingPath("motd-dynamic.sh"), MotdPath: getTestingPath("motd-static")}

	f, _ := os.Stat(getTestingPath("motd-dynamic.sh"))
	originalMode := f.Mode()
	defer os.Chmod(getTestingPath("motd-dynamic.sh"), originalMode)
	os.Chmod(getTestingPath("motd-dynamic.sh"), 0755)

	os.Stat(getTestingPath("motd-dynamic.sh"))

	output := readOutput(func() {
		printMotd(c)
	})

	if !strings.HasPrefix(output, "This is dynamic motd") {
		t.Error("TestMotdDynamic: result does not match expected value")
	}
}

func TestMotdStatic(t *testing.T) {
	c := &config{PrintMotd: true, DynamicMotd: false, DynamicMotdPath: getTestingPath("motd-dynamic.sh"), MotdPath: getTestingPath("motd-static"), DaemonMode: true}

	output := readOutput(func() {
		printMotd(c)
	})

	if !strings.HasPrefix(output, "This is static motd.") {
		t.Error("TestMotdStatic: result does not match expected value")
	}
}

func TestMotdStaticEmpty(t *testing.T) {
	c := &config{PrintMotd: true, DynamicMotd: false, DynamicMotdPath: getTestingPath("motd-dynamic.sh"), MotdPath: getTestingPath("motd-static-empty")}

	output := readOutput(func() {
		printMotd(c)
	})

	if output != "" {
		t.Error("TestMotdStaticEmpty: result does not match expected value")
	}
}

func TestMotdDefault(t *testing.T) {
	c := &config{PrintMotd: true}

	output := readOutput(func() {
		printMotd(c)
	})

	if !strings.HasPrefix(output, defaultMotd) {
		t.Error("TestMotdDefault: result does not match expected value")
	}
}

func TestRevertColorEscaping(t *testing.T) {
	if revertColorEscaping("") != "" {
		t.Error("TestRevertColorEscaping: there should not be nothing to be handled")
	}

	value := "\\033[0;m\\x1b[0;m"
	expected := "\033[0;m\x1b[0;m"

	if revertColorEscaping(value) != expected {
		t.Error("TestRevertColorEscaping: result does not match expected value")
	}
}

func TestSetColors(t *testing.T) {
	output := readOutput(func() {
		resetColors()
	})
	if output != "\x1b[0;0m\n" {
		t.Error("TestSetColors: result does not match to resetting value")
	}

	output = readOutput(func() {
		setColors("31", "")
	})
	if output != "\x1b[0;31m\n" {
		t.Error("TestSetColors: result does not match to defined foreground value")
	}

	output = readOutput(func() {
		setColors("", "41")
	})
	if output != "\x1b[0;41m\n" {
		t.Error("TestSetColors: result does not match to defined background value")
	}

	output = readOutput(func() {
		setColors("31", "41")
	})
	if output != "\x1b[0;31;41m\n" {
		t.Error("TestSetColors: result does not match to defined foreground and background value")
	}

}
