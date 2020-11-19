package main

// #include <stdlib.h>
// #include <utmp.h>
// #include <utmpx.h>
import "C"
import (
	"log"
	"os"
	"time"
)

// Adds UTMPx entry as user process
func addUtmpEntry(username string, pid int, ttyNo string) *C.struct_utmpx {
	utmp := &C.struct_utmpx{}
	xdisplay := os.Getenv(envDisplay)

	utmp.ut_type = C.USER_PROCESS
	utmp.ut_pid = C.int(pid)
	utmp.ut_line = strToC32Char("tty" + ttyNo)
	if xdisplay != "" {
		utmp.ut_id = strToC4Char(xdisplay)
	} else {
		utmp.ut_id = strToC4Char(ttyNo)
	}
	utmp.ut_tv.tv_sec = C.int(int(time.Now().Unix()))
	utmp.ut_user = strToC32Char(username)
	utmp.ut_host = strToC256Char(xdisplay)

	putUtmpEntry(utmp)

	return utmp
}

// End UTMPx entry by marking as dead process
func endUtmpEntry(utmp *C.struct_utmpx) {
	utmp.ut_type = C.DEAD_PROCESS
	utmp.ut_tv.tv_sec = C.int(int(time.Now().Unix()))

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

	C.updwtmp(C.CString("/var/log/wtmp"), utmp)
}

// Converts string to [4]C.char
func strToC4Char(data string) [4]C.char {
	result := [4]C.char{}

	for i := 0; i < len(data) && i < 4; i++ {
		result[i] = C.char(data[i])
	}
	return result
}

// Converts string to [32]C.char
func strToC32Char(data string) [32]C.char {
	result := [32]C.char{}

	for i := 0; i < len(data) && i < 32; i++ {
		result[i] = C.char(data[i])
	}
	return result
}

// Converts string to [256]C.char
func strToC256Char(data string) [256]C.char {
	result := [256]C.char{}

	for i := 0; i < len(data) && i < 256; i++ {
		result[i] = C.char(data[i])
	}
	return result
}
