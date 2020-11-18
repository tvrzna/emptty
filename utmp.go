package main

// #include <stdlib.h>
// #include <utmp.h>
import "C"
import (
	"time"
)

// Adds UTMP entry as user process
func addUtmpEntry(username string, displayPid int, ttyNo string) *C.struct_utmp {
	utmp := &C.struct_utmp{}
	utmp.ut_type = C.USER_PROCESS
	utmp.ut_pid = C.int(displayPid)
	utmp.ut_line = strToC32Char("tty" + ttyNo)
	utmp.ut_id = strToC4Char(ttyNo)
	utmp.ut_tv.tv_sec = C.int(int(time.Now().Unix()))
	utmp.ut_user = strToC32Char(username)
	utmp.ut_host = strToC256Char("")
	utmp.ut_addr_v6 = [4]C.int{0, 0, 0, 0}

	C.setutent()
	C.pututline(utmp)

	return utmp
}

// End UTMP entry by marking as dead process
func endUtmpEntry(utmp *C.struct_utmp) {
	utmp.ut_type = C.DEAD_PROCESS
	utmp.ut_user = strToC32Char("LOGIN")
	C.setutent()
	C.pututline(utmp)
	C.endutent()
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
