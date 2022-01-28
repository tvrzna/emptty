//go:build noxlib
// +build noxlib

package src

import (
	"time"
)

const tagXlib = "noxlib"

type xdisplay struct {
	disp     string
	dispName string
}

// Slows down start by waiting to create X lock file
func (c *xdisplay) openXDisplay() error {
	for i := 0; i < 50; i++ {
		if fileExists("/tmp/.X11-unix/X" + c.dispName[1:]) {
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil
}
