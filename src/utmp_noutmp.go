// +build noutmp

package src

const tagUtmp = "noutmp"

// Adds UTMP entry as user process
func addUtmpEntry(username string, pid int, ttyNo string, xdisplay string) bool {
	return false
}

// End UTMP entry by marking as dead process
func endUtmpEntry(value bool) {
	// Nothing to do here
}

// Adds BTMP entry to log unsuccessful login attempt.
func addBtmpEntry(username string, pid int, ttyNo string) {
	// Nothing to do here
}
