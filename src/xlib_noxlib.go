// +build noxlib

package main

import (
	"os"
	"time"
)

type xdisplay struct {
	disp string
}

// Slows down start by waiting to create X lock file
func (c *xdisplay) openXDisplay() error {
	c.disp = os.Getenv(envDisplay)[1:]

	for i := 0; i < 50; i++ {
		if fileExists("/tmp/.X11-unix/X" + c.disp) {
			break
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil
}

// Nothing to do here
func (c *xdisplay) closeXDisplay() error {
	// Nothing to do here
	return nil
}
