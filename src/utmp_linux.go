package src

// #include <paths.h>
// #include <utmp.h>
// #include <utmpx.h>
import "C"

// Puts UTMP entry into wtmp file
func updwtmpx(utmpx *C.struct_utmpx) {
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

	C.updwtmp(C.CString(C._PATH_WTMP), utmp)
}
