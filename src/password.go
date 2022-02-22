package src

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
)

// Reads password without echoing it
func readPassword() (string, error) {
	fd := os.Stdout.Fd()

	c := makeInterruptChannel()

	go handlePasswordInterrupt(c, fd)

	err := setTerminalEcho(fd, false)
	if err != nil {
		return "", err
	}
	defer signal.Stop(c)
	defer setTerminalEcho(fd, true)

	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	fmt.Println()
	return input[:len(input)-1], nil
}

// Enables echo on interruption and provide interrupt.
func handlePasswordInterrupt(c chan os.Signal, fd uintptr) {
	<-c
	setTerminalEcho(fd, true)
	os.Exit(-1)
}
