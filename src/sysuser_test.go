package src

import (
	"os/user"
	"testing"
)

func TestGetSysuser(t *testing.T) {
	usr := &user.User{Uid: "3000", Gid: "2000", Username: "Dummy", Name: "There is no name", HomeDir: "/dev/null"}

	u := getSysuser(usr)

	if u.strUid() != usr.Uid {
		t.Error("TestGetSysuser: uid does not match")
	}

	if u.strGid() != usr.Gid {
		t.Error("TestGetSysuser: gid does not match")
	}

	if u.uidu32() != 3000 {
		t.Error("TestGetSysuser: uid32 does not match")
	}

	if u.gidu32() != 2000 {
		t.Error("TestGetSysuser: gid32 does not match")
	}
}

func TestSysuserEnviron(t *testing.T) {
	u := &sysuser{}
	u.env = make(map[string]string)

	if len(u.environ()) != 0 {
		t.Error("TestSysuserEnviron: no environmental variable was expected")
	}

	u.setenv("", "value")
	if len(u.environ()) != 0 {
		t.Error("TestSysuserEnviron: inserted variable with empty name")
	}

	u.setenv("     ", "value")
	if len(u.environ()) != 0 {
		t.Error("TestSysuserEnviron: inserted variable with blank name")
	}

	if u.getenv("   ") != "" {
		t.Error("TestSysuserEnviron: variable with blank name could not be accessible")
	}

	if u.getenv("non-existent") != "" {
		t.Error("TestSysuserEnviron: found non-existent variable")
	}

	u.setenv("key", "value")
	if u.getenv("key") != "value" {
		t.Error("TestSysuserEnviron: environmental variable does not contain expected value")
	}

	if len(u.environ()) != 1 {
		t.Error("TestSysuserEnviron: 1 environmental variable was expected")
	}

	u.setenv("key", "value2")
	if u.getenv("key") == "value" {
		t.Error("TestSysuserEnviron: environmental variable is not being updated")
	}

	if len(u.environ()) != 1 {
		t.Error("TestSysuserEnviron: 1 environmental variable was expected after update")
	}

	u.setenv("key2", "value")
	if u.getenv("key") == "value" {
		t.Error("TestSysuserEnviron: environmental variable is not being updated")
	}

	if len(u.environ()) == 1 {
		t.Error("TestSysuserEnviron: 2 environmental variable were expected")
	}

	u.setenvIfEmpty("key3", "value1")
	if u.getenv("key3") != "value1" {
		t.Errorf("TestSysuserEnviron: key3 has unexpected value '%s'", u.getenv("key3"))
	}

	u.setenvIfEmpty("key3", "value2")
	if u.getenv("key3") != "value1" {
		t.Errorf("TestSysuserEnviron: key3 has unexpected value '%s'", u.getenv("key3"))
	}

}
