package src

// #include <pwd.h>
import "C"
import (
	"os/user"
	"strconv"
	"strings"
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
	u := &sysuser{}

	u.username = usr.Username
	u.homedir = usr.HomeDir
	u.uid, _ = strconv.Atoi(usr.Uid)
	u.gid, _ = strconv.Atoi(usr.Gid)
	u.env = make(map[string]string)

	if strGids, err := usr.GroupIds(); err == nil {
		for _, val := range strGids {
			value, _ := strconv.Atoi(val)
			u.gids = append(u.gids, int(value))
			u.gidsu32 = append(u.gidsu32, uint32(value))
		}
	}

	return u
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
	if pwd := C.getpwuid(C.uint(u.uid)); pwd != nil && pwd.pw_shell != nil {
		return C.GoString(pwd.pw_shell)
	}
	return "/bin/sh"
}
