package main

// #include <utmpx.h>
import "C"

// Puts UTMP entry into wtmp file
func updwtmpx(utmpx *C.struct_utmpx) {
	// Nothing to do here.
}
