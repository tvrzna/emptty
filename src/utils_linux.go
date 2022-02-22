package src

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	_KDSETLED = 0x4B32
	_KDGKBLED = 0x4B64
	_KDSKBLED = 0x4B65

	_K_SCROLLLOCK = 0x01
	_K_NUMLOCK    = 0x02
	_K_CAPSLOCK   = 0x04
)

// Sets fsuid, fsgid and fsgroups according sysuser
func setFsUser(usr *sysuser) {
	err := syscall.Setfsuid(usr.uid)
	handleErr(err)

	err = syscall.Setfsgid(usr.gid)
	handleErr(err)

	err = syscall.Setfsgid(usr.gid)
	handleErr(err)
}

// Sets keyboard LEDs
func setKeyboardLeds(tty *os.File, scrolllock, numlock, capslock bool) {
	var flags uint64

	// Read current keyboards flags
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(tty.Fd()), uintptr(_KDGKBLED), uintptr(unsafe.Pointer(&flags)))

	if scrolllock {
		flags |= _K_SCROLLLOCK
	}
	if numlock {
		flags |= _K_NUMLOCK
	}
	if capslock {
		flags |= _K_CAPSLOCK
	}

	if scrolllock || numlock || capslock {
		// Magic constant that allows user changes
		flags |= 0x30

		// Flags are used also for leds to keep flag valid to led
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(tty.Fd()), uintptr(_KDSKBLED), uintptr(flags))
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(tty.Fd()), uintptr(_KDSETLED), uintptr(flags))
	}
}

// Enables or disables echo depending on status
func setTerminalEcho(fd uintptr, status bool) error {
	var termios = &syscall.Termios{}

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}

	if status {
		termios.Lflag |= syscall.ECHO
	} else {
		termios.Lflag &^= syscall.ECHO
	}

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(termios))); err != 0 {
		return err
	}
	return nil
}
