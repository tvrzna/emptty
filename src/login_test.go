package src

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

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

func TestGetUptime(t *testing.T) {
	if getUptime() == 0 {
		t.Error("TestGetUptime: Expected non-zero value")
	}
}

func TestGetLoginRetryPath(t *testing.T) {
	for i := 0; i < 5; i++ {
		c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: 2, Tty: i}
		if !strings.HasSuffix(getLoginRetryPath(c), strconv.Itoa(i)) {
			t.Error("TestGetLoginRetryPath: Expected login retry path to match tty")
		}
	}
}

func TestReadWriteRetryFile(t *testing.T) {
	retryPath := "/tmp/emptty-test/retry"

	for i := 0; i < 5; i++ {
		writeRetryFile(retryPath, 4+i, float64(123.456+float64(i)))
		retries, time := readRetryFile(retryPath)

		// Need a little wiggle room for the float
		if retries != 4+i || time < float64(123+i) || time > float64(124+i) {
			t.Error("TestReadWriteRetryFile: Unexpected values returned")
		}
	}

	os.RemoveAll(retryPath)
}
