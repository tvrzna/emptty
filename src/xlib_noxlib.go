//go:build noxlib
// +build noxlib

package src

import (
	"time"
)

const tagXlib = "noxlib"

// Slows down start by waiting to create X lock file
func openXDisplay(dispName string) error {
	for i := 0; i < 50; i++ {
		if fileExists("/tmp/.X11-unix/X" + dispName[1:]) {
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil
}
