// +build !noxlib

package main

// #cgo LDFLAGS: -lX11
// #include <stdlib.h>
// #include <X11/Xlib.h>
import "C"
import (
	"errors"
	"time"
	"unsafe"
)

type xdisplay struct {
	disp *C.Display
}

// Opens XDisplay with xlib.
func (c *xdisplay) openXDisplay() error {
	if c.disp != nil {
		return errors.New("X Display is already opened")
	}
	for i := 0; i < 50; i++ {
		d := C.XOpenDisplay(nil)
		if d != nil {
			c.disp = d
			return nil
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return errors.New("Could not open X Display")
}

// Closes XDisplay with xlib
func (c *xdisplay) closeXDisplay() error {
	if c.disp == nil {
		return errors.New("Not connected to any X Display")
	}
	if C.XCloseDisplay(c.disp) == 0 {
		C.free(unsafe.Pointer(c.disp))
		return nil
	}
	return errors.New("Could not close active X Display")
}
