//go:build nopam
// +build nopam

package src

// #include <sys/types.h>
// #include <login_cap.h>
// #include <bsd_auth.h>
// #include <string.h>
// #include <stdlib.h>
import "C"
import "unsafe"

// Tries to authorize user with password.
func (n *nopamHandle) authPassword(username string, password string) bool {
	usr := C.CString(username)
	defer C.free(unsafe.Pointer(usr))

	passwd := C.CString(password)
	defer C.free(unsafe.Pointer(passwd))

	return C.auth_userokay(usr, nil, nil, passwd) > 0
}
