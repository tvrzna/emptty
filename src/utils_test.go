package src

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

// Testing method to get current working directory
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		logFatal(err)
	}
	return cwd
}

// Testing method to get path of file/directory in testing directory
func getTestingPath(path string) string {
	return getCwd() + "/../res/testing/" + path
}

// Testing method to steal os.Stdout and check printed output.
func readOutput(method func()) string {
	original := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	log.SetOutput(w)

	method()

	w.Close()
	output, _ := ioutil.ReadAll(r)
	os.Stdout = original
	log.SetOutput(original)

	if testing.Verbose() {
		os.Stdout.Write(output)
	}

	return string(output)
}

func TestConvertColor(t *testing.T) {
	if convertColor("UNKNOWN", true) != "" {
		t.Error("TestConvertColor: unexpected color for unknown")
	}

	// BLACK - 30
	if convertColor("BLACK", true) != "30" {
		t.Error("TestConvertColor: unexpected color for BLACK foreground")
	}
	if convertColor("BLACK", false) != "40" {
		t.Error("TestConvertColor: unexpected color for BLACK background")
	}
	if convertColor("LIGHT_BLACK", true) != "90" {
		t.Error("TestConvertColor: unexpected color for LIGHT BLACK foreground")
	}
	if convertColor("LIGHT_BLACK", false) != "100" {
		t.Error("TestConvertColor: unexpected color for LIGHT BLACK background")
	}

	// RED - 31
	if convertColor("RED", true) != "31" {
		t.Error("TestConvertColor: unexpected color for RED foreground")
	}
	if convertColor("RED", false) != "41" {
		t.Error("TestConvertColor: unexpected color for RED background")
	}
	if convertColor("LIGHT_RED", true) != "91" {
		t.Error("TestConvertColor: unexpected color for LIGHT RED foreground")
	}
	if convertColor("LIGHT_RED", false) != "101" {
		t.Error("TestConvertColor: unexpected color for LIGHT RED background")
	}

	// GREEN - 32
	if convertColor("GREEN", true) != "32" {
		t.Error("TestConvertColor: unexpected color for GREEN foreground")
	}
	if convertColor("GREEN", false) != "42" {
		t.Error("TestConvertColor: unexpected color for GREEN background")
	}
	if convertColor("LIGHT_GREEN", true) != "92" {
		t.Error("TestConvertColor: unexpected color for LIGHT GREEN foreground")
	}
	if convertColor("LIGHT_GREEN", false) != "102" {
		t.Error("TestConvertColor: unexpected color for LIGHT GREEN background")
	}

	// YELLOW - 33
	if convertColor("YELLOW", true) != "33" {
		t.Error("TestConvertColor: unexpected color for YELLOW foreground")
	}
	if convertColor("YELLOW", false) != "43" {
		t.Error("TestConvertColor: unexpected color for YELLOW background")
	}
	if convertColor("LIGHT_YELLOW", true) != "93" {
		t.Error("TestConvertColor: unexpected color for LIGHT YELLOW foreground")
	}
	if convertColor("LIGHT_YELLOW", false) != "103" {
		t.Error("TestConvertColor: unexpected color for LIGHT YELLOW background")
	}

	// BLUE - 34
	if convertColor("BLUE", true) != "34" {
		t.Error("TestConvertColor: unexpected color for BLUE foreground")
	}
	if convertColor("BLUE", false) != "44" {
		t.Error("TestConvertColor: unexpected color for BLUE background")
	}
	if convertColor("LIGHT_BLUE", true) != "94" {
		t.Error("TestConvertColor: unexpected color for LIGHT BLUE foreground")
	}
	if convertColor("LIGHT_BLUE", false) != "104" {
		t.Error("TestConvertColor: unexpected color for LIGHT BLUE background")
	}

	// MAGENTA - 35
	if convertColor("MAGENTA", true) != "35" {
		t.Error("TestConvertColor: unexpected color for MAGENTA foreground")
	}
	if convertColor("MAGENTA", false) != "45" {
		t.Error("TestConvertColor: unexpected color for MAGENTA background")
	}
	if convertColor("LIGHT_MAGENTA", true) != "95" {
		t.Error("TestConvertColor: unexpected color for LIGHT MAGENTA foreground")
	}
	if convertColor("LIGHT_MAGENTA", false) != "105" {
		t.Error("TestConvertColor: unexpected color for LIGHT MAGENTA background")
	}

	// CYAN - 36
	if convertColor("CYAN", true) != "36" {
		t.Error("TestConvertColor: unexpected color for CYAN foreground")
	}
	if convertColor("CYAN", false) != "46" {
		t.Error("TestConvertColor: unexpected color for CYAN background")
	}
	if convertColor("LIGHT_CYAN", true) != "96" {
		t.Error("TestConvertColor: unexpected color for LIGHT CYAN foreground")
	}
	if convertColor("LIGHT_CYAN", false) != "106" {
		t.Error("TestConvertColor: unexpected color for LIGHT CYAN background")
	}

	// WHITE - 37
	if convertColor("WHITE", true) != "37" {
		t.Error("TestConvertColor: unexpected color for WHITE foreground")
	}
	if convertColor("WHITE", false) != "47" {
		t.Error("TestConvertColor: unexpected color for WHITE background")
	}
	if convertColor("LIGHT_WHITE", true) != "97" {
		t.Error("TestConvertColor: unexpected color for LIGHT WHITE foreground")
	}
	if convertColor("LIGHT_WHITE", false) != "107" {
		t.Error("TestConvertColor: unexpected color for LIGHT WHITE background")
	}
}

func TestCmdAsUser(t *testing.T) {
	u := &sysuser{uid: 3000, gid: 2000}

	cmd := cmdAsUser(u, "/dev/null", "another", "and_another")

	if cmd.SysProcAttr.Credential.Uid != 3000 {
		t.Error("TestCmdAsUser: unexpected UID")
	}

	if cmd.SysProcAttr.Credential.Gid != 2000 {
		t.Error("TestCmdAsUser: unexpected UID")
	}
}

func TestSetKeyboardLeds(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "emptty-led-test")
	if err != nil {
		t.Error("TestSetKeyboardLeds: could not open test file")
	}

	setKeyboardLeds(f, true, true, true)
	setKeyboardLeds(f, false, false, false)

	f.Close()
	err = os.Remove(f.Name())
	if err != nil {
		t.Error("TestSetKeyboardLeds: could not remove test file")
	}
}
