package src

import (
	"bufio"
	"fmt"
	"os"
)

// Reads password without echoing it
func readPassword() (string, error) {
	fd := os.Stdout.Fd()

	if err := setTerminalEcho(fd, false); err != nil {
		return "", err
	}
	defer setTerminalEcho(fd, true)

	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	fmt.Println()
	return input[:len(input)-1], nil
}
