package main

// #include <sys/time.h>
// #include <utmpx.h>
import "C"

// Puts timeval data into UTMPx entry
func putTimeToUtmpEntry(utmp *C.struct_utmpx) {
	tv := &C.struct_timeval{}
	C.gettimeofday(tv, nil)
	utmp.ut_tv.tv_sec = tv.tv_sec
	utmp.ut_tv.tv_usec = tv.tv_usec
}
