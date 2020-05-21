package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// propertyFunc defines method to be invoked during readProperties method for each record.
type propertyFunc func(key string, value string) error

// readProperties reads defined filePath per line and parses each key-value pair.
// These pairs are used as parameters for invoking propertyFunc
func readProperties(filePath string, method propertyFunc) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.New("Could not open file " + filePath)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && strings.Index(line, "=") >= 0 {
			splitIndex := strings.Index(line, "=")
			key := strings.ReplaceAll(line[:splitIndex], "export ", "")
			value := line[splitIndex+1:]
			if strings.Index(value, "#") >= 0 {
				value = value[:strings.Index(value, "#")]
			}
			err := method(key, value)
			handleErr(err)
		}
	}
	return scanner.Err()
}

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
