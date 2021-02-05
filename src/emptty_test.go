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
