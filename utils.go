package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

// Perform switch to defined TTY.
func switchTTY(ttyNumber int) {
	if ttyNumber > 0 {
		ttyCmd := exec.Command("/usr/bin/chvt", strconv.Itoa(ttyNumber))
		ttyCmd.Run()
	}
}

// If error is not nil, otherwise it prints error, waits for user input and then exits the program.
func handleErr(err error) {
	if err != nil {
		log.Print(err)
		fmt.Printf("\nPress Enter to continue...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(1)
	}
}
