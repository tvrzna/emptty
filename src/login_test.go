package src

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

type MockedRetryPathProvider struct {
	fileName string
}

func (r *MockedRetryPathProvider) getLoginRetryPath(conf *config) string {
	if r.fileName == "" {
		f, _ := os.CreateTemp(os.TempDir(), "emptty-login-retry-"+conf.strTTY())
		r.fileName = f.Name()
		f.Close()
	}
	return r.fileName
}

func TestHandleLoginRetriesInfinite(t *testing.T) {
	r := &MockedRetryPathProvider{}
	c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: -1}

	for i := 0; i < 5; i++ {
		err := handleLoginRetries(c, r)
		if err != nil {
			t.Error("TestHandleLoginRetriesInfinite: No error from handleLoginRetries was expected")
		}
	}

	os.Remove(r.getLoginRetryPath(c))
}

func TestHandleLoginRetriesNoRetry(t *testing.T) {
	r := &MockedRetryPathProvider{}
	c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: 0}

	for i := 0; i < 5; i++ {
		err := handleLoginRetries(c, r)
		if err != nil {
			break
		}
		if i > 0 {
			t.Error("TestHandleLoginRetriesNoRetry: No retry was expected")
		}
	}

	os.Remove(r.getLoginRetryPath(c))
}

func TestHandleLoginRetries2Retries(t *testing.T) {
	r := &MockedRetryPathProvider{}
	c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: 2}

	for i := 0; i < 5; i++ {
		err := handleLoginRetries(c, r)
		if err != nil {
			break
		}
		if i > 3 {
			t.Error("TestHandleLoginRetries2Retries: No retry was expected")
		}
	}

	os.Remove(r.getLoginRetryPath(c))
}

func TestGetUptime(t *testing.T) {
	if getUptime() == 0 {
		t.Error("TestGetUptime: Expected non-zero value")
	}
}

func TestGetLoginRetryPath(t *testing.T) {
	r := &DefaultLoginRetryPathProvider{}
	for i := 0; i < 5; i++ {
		c := &config{Autologin: true, AutologinSession: "/dev/null", AutologinMaxRetry: 2, Tty: i}
		if !strings.HasSuffix(r.getLoginRetryPath(c), strconv.Itoa(i)) {
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
