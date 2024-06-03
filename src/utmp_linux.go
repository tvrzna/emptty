//go:build !noutmp

package src

// #include <paths.h>
// #include <stdlib.h>
// #include <utmp.h>
// #include <utmpx.h>
import "C"
import "unsafe"

// Converts UTMPx entry into UTMP structure.
func convertUtmpxToUtmp(utmpx *C.struct_utmpx) *C.struct_utmp {
	utmp := &C.struct_utmp{}
	utmp.ut_type = utmpx.ut_type
	utmp.ut_pid = utmpx.ut_pid
	utmp.ut_line = utmpx.ut_line
	utmp.ut_id = utmpx.ut_id
	utmp.ut_tv.tv_sec = utmpx.ut_tv.tv_sec
	utmp.ut_tv.tv_usec = utmpx.ut_tv.tv_usec
	utmp.ut_user = utmpx.ut_user
	utmp.ut_host = utmpx.ut_host
	utmp.ut_addr_v6 = utmpx.ut_addr_v6

	return utmp
}

// Puts UTMP entry into wtmp file.
func updwtmpx(utmpx *C.struct_utmpx) {
	wtmpPath := C.CString(C._PATH_WTMP)
	C.updwtmp(wtmpPath, convertUtmpxToUtmp(utmpx))
	C.free(unsafe.Pointer(wtmpPath))
}

// Adds BTMP entry to log unsuccessful login attempt.
func addBtmpEntry(username string, pid int, ttyNo string) {
	btmpPath := C.CString("/var/log/btmp")
	utmpx := prepareUtmpEntry(username, pid, ttyNo, "")
	C.updwtmp(btmpPath, convertUtmpxToUtmp(utmpx))
	C.free(unsafe.Pointer(btmpPath))
}
