package src

import (
	"bytes"
	"testing"
)

func TestPrintIssue(t *testing.T) {
	readOutput(func() {
		printIssue(getTestingPath("issue"), "X")
	})
}

func TestPrintIssueFull(t *testing.T) {
	readOutput(func() {
		printIssue(getTestingPath("issue_anything"), "X")
	})
}

func TestPrintIssueWithMoreLines(t *testing.T) {
	output := readOutput(func() {
		printIssue(getTestingPath("issue_more_new_lines"), "X")
	})

	if output != "Hello with new lines\n" {
		t.Error("TestPrintIssueWithMoreLines: issue_more_new_lines after being parsed does not equals to expected value.")
	}
}

func TestClearScreen(t *testing.T) {
	output := readOutput(func() {
		clearScreen(nil)
	})

	if output != strCleanScreen {
		t.Error("TestClearScreen: screen was not cleared")
	}
}

func TestClearScreenWithOutput(t *testing.T) {
	buf := new(bytes.Buffer)

	clearScreen(buf)

	if buf.String() != strCleanScreen {
		t.Error("TestClearScreenWithOutput: screen was not cleared")
	}
}

func TestSwitchTTY(t *testing.T) {
	conf := &config{}
	conf.SwitchTTY = false
	conf.Tty = -99

	if switchTTY(conf) {
		t.Error("TestSwitchTTY: attempt to switch tty, even it is disabled")
	}

	conf.SwitchTTY = true

	if switchTTY(conf) {
		t.Error("TestSwitchTTY: attempt to switch tty with negative number")
	}
}
