package src

import (
	"testing"
)

func TestPrintDefaultMotd(t *testing.T) {
	output := readOutput(func() {
		printDefaultMotd()
	})

	if output != default_motd+"\n\n" {
		t.Error("TestPrintDefaultMotd: default motd does not match")
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
		t.Error("TestSetColors: result does not match to reseting value")
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
