package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	default_motd = `┌─┐┌┬┐┌─┐┌┬┐┌┬┐┬ ┬
├┤ │││├─┘ │  │ └┬┘
└─┘┴ ┴┴   ┴  ┴  ┴   ` + version

	pathMotd  = "/etc/emptty/motd"
	pathIssue = "/etc/issue"
)

// Prints motd, if pathMotd exists, prints it; otherwise it prints default motd.
func printMotd() {
	if fileExists(pathMotd) {
		file, err := os.Open(pathMotd)
		defer file.Close()
		if err != nil {
			printDefaultMotd()
		}
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			fmt.Println(revertColorEscaping(scan.Text()))
		}
		// Clear to default
		fmt.Print("\x1b[0m\n")
		return
	} else {
		printDefaultMotd()
	}
}

// Prints default motd.
func printDefaultMotd() {
	fmt.Printf("%s\n\n", default_motd)
}

// Reverts escaped color definitions to real color values.
func revertColorEscaping(value string) string {
	if value != "" {
		result := strings.ReplaceAll(value, "\\x1b", "\x1b")
		result = strings.ReplaceAll(result, "\\033", "\x1b")
		return result
	}
	return value
}
