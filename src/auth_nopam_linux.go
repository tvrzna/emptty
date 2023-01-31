//go:build nopam
// +build nopam

package src

// #include <pwd.h>
// #include <crypt.h>
// #include <shadow.h>
// #include <string.h>
// #include <stdlib.h>
// #cgo linux LDFLAGS: -lcrypt
import "C"
import "unsafe"

// Tries to authorize user with password.
func authPassword(username string, password string) bool {
	usr := C.CString(username)
	defer C.free(unsafe.Pointer(usr))

	passwd := C.CString(password)
	defer C.free(unsafe.Pointer(passwd))

	var passhash *C.char

	if pwd := C.getspnam(usr); pwd != nil {
		passhash = pwd.sp_pwdp
	}

	if passhash == nil {
		if pwd := C.getpwnam(usr); pwd != nil {
			passhash = pwd.pw_passwd
		}
	}

	if passhash == nil {
		return false
	}

	encrypted := C.crypt(passwd, passhash)
	if encrypted == nil || C.strcmp(encrypted, passhash) != 0 {
		return false
	}
	return true
}
