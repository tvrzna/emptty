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
	cClose := make(chan int)

	go handlePasswordInterrupt(c, cClose, fd)

	err := setTerminalEcho(fd, false)
	if err != nil {
		return "", err
	}
	defer signal.Stop(c)
	defer func() { cClose <- 0 }()
	defer setTerminalEcho(fd, true)

	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	fmt.Println()
	return input[:len(input)-1], nil
}

// Enables echo on interruption and provide interrupt.
func handlePasswordInterrupt(c chan os.Signal, cClose chan int, fd uintptr) {
	select {
	case <-c:
		setTerminalEcho(fd, true)
		os.Exit(-1)
	case <-cClose:
		// nothing to do
	}
}
