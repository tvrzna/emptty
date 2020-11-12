package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	default_motd = `┌─┐┌┬┐┌─┐┌┬┐┌┬┐┬ ┬
├┤ │││├─┘ │  │ └┬┘
└─┘┴ ┴┴   ┴  ┴  ┴   ` + version

	pathDynamicMotd = "/etc/emptty/motd-gen.sh"
	pathMotd        = "/etc/emptty/motd"
	pathIssue       = "/etc/issue"
)

// Prints dynamic motd, if configured; otherwise prints motd, if pathMotd exists; otherwise it prints default motd.
func printMotd(conf *config) {
	if conf.dynamicMotd && fileIsExecutable(pathDynamicMotd) {
		cmd := exec.Command(pathDynamicMotd)
		dynamicMotd, err := cmd.Output()
		if err != nil {
			log.Print(err)
			printDefaultMotd()
			return
		}
		fmt.Print(revertColorEscaping(string(dynamicMotd)))
		resetColors()
	} else if fileExists(pathMotd) {
		file, err := os.Open(pathMotd)
		defer file.Close()
		if err != nil {
			log.Print(err)
			printDefaultMotd()
			return
		}
		scan := bufio.NewScanner(file)
		for scan.Scan() {
			fmt.Println(revertColorEscaping(scan.Text()))
		}
		resetColors()
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

// Resets colors to default.
func resetColors() {
	fmt.Print("\x1b[0m\n")
}
