package src

import (
	"os"
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
