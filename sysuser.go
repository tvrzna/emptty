package main

import (
	"os/user"
	"strconv"
)

// Type sysuser defines default structure of user to easier passing of all values.
type sysuser struct {
	username string
	homedir  string
	uid      int
	gid      int
	gids     []int
	gidsu32  []uint32
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
