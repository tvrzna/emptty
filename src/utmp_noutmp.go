// +build noutmp

package src

// Adds UTMP entry as user process
func addUtmpEntry(username string, pid int, ttyNo string, xdisplay string) bool {
	return false
}

// End UTMP entry by marking as dead process
func endUtmpEntry(value bool) {
	// Nothing to do here
}
