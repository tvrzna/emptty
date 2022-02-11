package src

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestParseLogging(t *testing.T) {
	var logging enLogging

	logging = parseLogging("", constLogDefault)
	if logging != Default {
		t.Error("TestParseLogging: wrong default value")
	}

	logging = parseLogging(constLogDefault, constLogDefault)
	if logging != Default {
		t.Error("TestParseLogging: wrong parsed value for default")
	}

	logging = parseLogging(constLogAppending, constLogDefault)
	if logging != Appending {
		t.Error("TestParseLogging: wrong parsed value for appending")
	}

	logging = parseLogging(constLogDisabled, constLogDefault)
	if logging != Disabled {
		t.Error("TestParseLogging: wrong parsed value for disabled")
	}

	logging = parseLogging("aaa", "bbb")
	if logging != Default {
		t.Error("TestParseLogging: wrong fallback value")
	}
}

func TestInitSessionErrorLogger(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "emptty-session-log-file")
	fileName := f.Name()
	f.Close()

	conf := &config{SessionErrLogFile: f.Name(), SessionErrLog: Default}
	sessFile, sessErr := initSessionErrorLogger(conf)
	sessFile.Close()
	os.Remove(fileName + pathLogFileOldSuffix)
	os.Remove(fileName)
	if sessErr != nil {
		t.Error("TestInitSessionErrorLogger: unexpected error", sessErr)
	}

	conf.SessionErrLog = Appending
	sessFile, sessErr = initSessionErrorLogger(conf)
	sessFile.Close()
	if sessErr != nil {
		t.Error("TestInitSessionErrorLogger: unexpected error", sessErr)
	}

	conf.SessionErrLog = Disabled
	sessFile, sessErr = initSessionErrorLogger(conf)
	sessFile.Close()
	os.Remove(fileName)
	if sessErr != nil {
		t.Error("TestInitSessionErrorLogger: unexpected error", sessErr)
	}
}

func TestInitLogger(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "emptty-log-file.[TTY_NUMBER]")
	fileName := f.Name()
	f.Close()

	conf := &config{LoggingFile: f.Name(), Logging: Default}
	initLogger(conf)
	os.Remove(fileName + pathLogFileOldSuffix)
	os.Remove(fileName)

	conf.Logging = Appending
	initLogger(conf)

	conf.Logging = Disabled
	initLogger(conf)
	os.Remove(fileName)
}

func TestLogPrint(t *testing.T) {
	expected := "expected message"
	output := readOutput(func() {
		logPrint(expected)
	})

	if !strings.Contains(output, expected) {
		t.Errorf("TestLogPrint: '%s' was expected, but was '%s'", expected, output)
	}
}

func TestLogPrintf(t *testing.T) {
	expected := "expected message"
	output := readOutput(func() {
		logPrintf("expected %s", "message")
	})

	if !strings.Contains(output, expected) {
		t.Errorf("TestLogPrint: '%s' was expected, but was '%s'", expected, output)
	}
}

func TestHandleErr(t *testing.T) {
	TEST_MODE = true

	output := readOutput(func() {
		handleErr(nil)
	})
	if output != "" {
		t.Errorf("TestHandleErr: output should have been empty, but was '%s'", output)
	}

	output = readOutput(func() {
		handleErr(errors.New("THIS IS ERROR"))
	})
	if !strings.Contains(output, "THIS IS ERROR") {
		t.Errorf("TestHandleErr: 'THIS IS ERROR' was expected, but was '%s'", output)
	}
}

func TestHandleStrErr(t *testing.T) {
	TEST_MODE = true

	output := readOutput(func() {
		handleStrErr("")
	})
	if output != "" {
		t.Errorf("TestHandleStrErr: output should have been empty, but was '%s'", output)
	}

	output = readOutput(func() {
		handleStrErr("THIS IS ERROR")
	})
	if !strings.Contains(output, "THIS IS ERROR") {
		t.Errorf("TestHandleStrErr: 'THIS IS ERROR' was expected, but was '%s'", output)
	}
}

func TestBackupFileIfNotFolder(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "emptty-data")
	fileName := f.Name()
	f.Close()

	backupFileIfNotFolder(fileName + "/file")
	backupFileIfNotFolder(fileName + "/file")

	os.Remove(fileName)
}
