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

	pwd := C.getspnam(usr)
	if pwd != nil {
		passhash = pwd.sp_pwdp
	}

	if passhash == nil {
		pwd := C.getpwnam(usr)

		if pwd != nil {
			passhash = pwd.pw_passwd
		}
	}

	if passhash == nil {
		return false
	}

	crypted := C.crypt(passwd, passhash)
	if C.strcmp(crypted, passhash) != 0 {
		return false
	}
	return true
}
