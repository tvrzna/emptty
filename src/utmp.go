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
	"log"
	"os"
	"unsafe"
)

// Adds UTMPx entry as user process
func addUtmpEntry(username string, pid int, ttyNo string) *C.struct_utmpx {
	utmp := &C.struct_utmpx{}
	xdisplay := os.Getenv(envDisplay)

	id := xdisplay
	if id == "" {
		id = ttyNo
	}

	ut_pid := C.int(pid)
	ut_id := C.CString(id)
	ut_line := C.CString("tty" + ttyNo)
	ut_user := C.CString(username)
	ut_host := C.CString(xdisplay)

	utmp.ut_type = C.USER_PROCESS
	C.prepareUtmpEntry(utmp, ut_pid, ut_id, ut_line, ut_user, ut_host)

	C.free(unsafe.Pointer(ut_id))
	C.free(unsafe.Pointer(ut_line))
	C.free(unsafe.Pointer(ut_user))
	C.free(unsafe.Pointer(ut_host))

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
		log.Println("Could not write into utmp.")
	}
	C.endutxent()

	updwtmpx(utmp)
}
