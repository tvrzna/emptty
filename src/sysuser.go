package src

import (
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

const (
	pathUserRetryFile = "/.cache/emptty/login-retry"
)

// Type sysuser defines default structure of user to easier passing of all values.
type sysuser struct {
	username string
	homedir  string
	uid      int
	gid      int
	gids     []int
	gidsu32  []uint32
	env      map[string]string
}

// Loads all necessary info about user into sysuser struct.
func getSysuser(usr *user.User) *sysuser {
	var result sysuser

	uid, _ := strconv.ParseInt(usr.Uid, 10, 32)
	gid, _ := strconv.ParseInt(usr.Gid, 10, 32)
	var gidsu32 []uint32
	var gids []int
	strGids, err := usr.GroupIds()
	if err == nil {
		for _, val := range strGids {
			value, _ := strconv.ParseInt(val, 10, 32)
			gids = append(gids, int(value))
			gidsu32 = append(gidsu32, uint32(value))
		}
	}

	result.username = usr.Username
	result.homedir = usr.HomeDir
	result.uid = int(uid)
	result.gid = int(gid)
	result.gids = gids
	result.gidsu32 = gidsu32
	result.env = make(map[string]string)

	return &result
}

// returns uid as uint32.
func (u *sysuser) uidu32() uint32 {
	return uint32(u.uid)
}

// returns gid as uint32.
func (u *sysuser) gidu32() uint32 {
	return uint32(u.gid)
}

// returns uid as string.
func (u *sysuser) strUid() string {
	return strconv.Itoa(u.uid)
}

// returns gid as string.
func (u *sysuser) strGid() string {
	return strconv.Itoa(u.gid)
}

// gets user's environmental variable by key.
func (u *sysuser) getenv(key string) string {
	if strings.TrimSpace(key) == "" {
		return ""
	}
	return u.env[key]
}

// sets user's environmental variable.
func (u *sysuser) setenv(key, value string) {
	if strings.TrimSpace(key) != "" {
		u.env[strings.TrimSpace(key)] = value
	}
}

// sets user's environmental variable only if is not already defined with same key
func (u *sysuser) setenvIfEmpty(key, value string) {
	if strings.TrimSpace(u.getenv(key)) == "" {
		u.setenv(key, value)
	}
}

// returns a copy of environmental variables.
func (u *sysuser) environ() []string {
	var result []string
	for key, value := range u.env {
		result = append(result, key+"="+value)
	}
	return result
}

// Reads default shell of user.
func (u *sysuser) getShell() string {
	out, err := exec.Command("/usr/bin/getent", "passwd", u.strUid()).Output()
	handleErr(err)

	ent := strings.Split(strings.TrimSuffix(string(out), "\n"), ":")
	return ent[6]
}

// Gets path to login retry file
func (u *sysuser) getLoginRetryPath() string {
	return u.homedir + pathUserRetryFile
}
