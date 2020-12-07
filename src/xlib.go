// +build !noxlib

package src

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
	disp     *C.Display
	dispName string
}

// Opens XDisplay with xlib.
func (c *xdisplay) openXDisplay() error {
	if c.disp != nil {
		return errors.New("X Display is already opened")
	}
	displayName := C.CString(c.dispName)
	defer C.free(unsafe.Pointer(displayName))
	for i := 0; i < 50; i++ {
		d := C.XOpenDisplay(displayName)
		if d != nil {
			c.disp = d
			return nil
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return errors.New("Could not open X Display")
}
