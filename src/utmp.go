// +build !noutmp
// +build !openbsd

package src

/*
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <utmpx.h>

void putTimeToUtmpEntry(struct utmpx * utmp) {
	struct timeval tv;
	gettimeofday (&tv, NULL);
	utmp->ut_tv.tv_sec = tv.tv_sec;
	utmp->ut_tv.tv_usec = tv.tv_usec;
}

void prepareUtmpEntry(struct utmpx * utmp, int pid, char* id, char* line, char* username, char* host) {
	utmp->ut_pid = pid;
	strncpy (utmp->ut_id, id, sizeof (utmp->ut_id));
	strncpy (utmp->ut_line, line, sizeof (utmp->ut_line));
	strncpy (utmp->ut_user, username, sizeof (utmp->ut_user));
	strncpy (utmp->ut_host, host, sizeof (utmp->ut_host));

	putTimeToUtmpEntry(utmp);
}
*/
import "C"
import (
	"unsafe"
)

const tagUtmp = ""

// Prepares UTMPx entry
func prepareUtmpEntry(username string, pid int, ttyNo string, xdisplay string) *C.struct_utmpx {
	utmp := &C.struct_utmpx{}

	id := xdisplay
	if id == "" {
		id = ttyNo
	}

	utPid := C.int(pid)
	utId := C.CString(id)
	utLine := C.CString("tty" + ttyNo)
	utUser := C.CString(username)
	utHost := C.CString(xdisplay)

	utmp.ut_type = C.USER_PROCESS
	C.prepareUtmpEntry(utmp, utPid, utId, utLine, utUser, utHost)

	C.free(unsafe.Pointer(utId))
	C.free(unsafe.Pointer(utLine))
	C.free(unsafe.Pointer(utUser))
	C.free(unsafe.Pointer(utHost))

	return utmp
}

// Adds UTMPx entry as user process
func addUtmpEntry(username string, pid int, ttyNo string, xdisplay string) *C.struct_utmpx {
	utmp := prepareUtmpEntry(username, pid, ttyNo, xdisplay)

	putUtmpEntry(utmp)

	return utmp
}

// End UTMPx entry by marking as dead process
func endUtmpEntry(utmp *C.struct_utmpx) {
	utmp.ut_type = C.DEAD_PROCESS
	C.putTimeToUtmpEntry(utmp)

	putUtmpEntry(utmp)
}

// Puts UTMPx entry into utmp file
func putUtmpEntry(utmp *C.struct_utmpx) {
	C.setutxent()
	if C.pututxline(utmp) == nil {
		logPrint("Could not write into utmp.")
	}
	C.endutxent()

	updwtmpx(utmp)
}
