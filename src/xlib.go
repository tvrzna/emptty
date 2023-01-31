//go:build !noxlib
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

const tagXlib = ""

// Opens XDisplay with xlib.
func openXDisplay(dispName string) error {
	displayName := C.CString(dispName)
	defer C.free(unsafe.Pointer(displayName))
	for i := 0; i < 50; i++ {
		d := C.XOpenDisplay(displayName)
		if d != nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("could not open X Display")
}
