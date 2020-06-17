package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const (
	strCleanScreen = "\x1b[H\x1b[2J"
)

// Starts emptty as daemon spawning emptty on defined TTY
func startDaemon() {
	conf := loadConfig()

	fTTY, err := os.OpenFile("/dev/tty"+conf.strTTY(), os.O_RDWR, 0700)
	if err != nil {
		log.Fatal(err)
	}

	clearScreen(fTTY)

	os.Stdout = fTTY
	os.Stderr = fTTY
	os.Stdin = fTTY

	switchTTY(conf)

	showLoginScreen(conf)

	clearScreen(fTTY)
}

// Clears terminal screen
func clearScreen(w *os.File) {
	if w == nil {
		fmt.Print(strCleanScreen)
	} else {
		w.Write([]byte(strCleanScreen))
	}
}

// Perform switch to defined TTY, if switchTTY is true and tty is greater than 0.
func switchTTY(conf *config) {
	if conf.switchTTY && conf.tty > 0 {
		ttyCmd := exec.Command("/usr/bin/chvt", conf.strTTY())
		ttyCmd.Run()
	}
}
