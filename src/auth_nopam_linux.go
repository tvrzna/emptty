//go:build nopam
// +build nopam

package src

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

	pwd := C.getspnam(usr)
	if pwd == nil {
		return false
	}
	crypted := C.crypt(passwd, pwd.sp_pwdp)
	if C.strcmp(crypted, pwd.sp_pwdp) != 0 {
		return false
	}
	return true
}
