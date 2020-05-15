package main

import (
	"log"
	"os/exec"
	"strconv"
)

// Perform switch to defined TTY.
func switchTTY(ttyNumber int) {
	if ttyNumber > 0 {
		ttyCmd := exec.Command("/bin/chvt", strconv.Itoa(ttyNumber))
		ttyCmd.Run()
	}
}

// If error is not nil, then use log.Fatal to stop application
func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
