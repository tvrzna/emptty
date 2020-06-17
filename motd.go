package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	default_motd = `┌─┐┌┬┐┌─┐┌┬┐┌┬┐┬ ┬
├┤ │││├─┘ │  │ └┬┘
└─┘┴ ┴┴   ┴  ┴  ┴   ` + version

	pathMotd  = "/etc/emptty/motd"
	pathIssue = "/etc/issue"
)

type issueVariable struct {
	value   string
	command []string
}

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

// Prints getty issue
func printIssue() {
	if fileExists(pathIssue) {
		bIssue, err := ioutil.ReadFile(pathIssue)
		issue := string(bIssue)
		if err == nil {
			vars := []issueVariable{
				issueVariable{"\\d", []string{"/usr/bin/date"}},
				issueVariable{"\\l", []string{"/usr/bin/ps", "-p", strconv.Itoa(os.Getpid()), "-o", "tty", "--no-headers"}},
				issueVariable{"\\m", []string{"/usr/bin/uname", "-m"}},
				issueVariable{"\\n", []string{"/usr/bin/uname", "-n"}},
				issueVariable{"\\r", []string{"/usr/bin/uname", "-r"}},
				issueVariable{"\\s", []string{"/usr/bin/uname", "-s"}},
				issueVariable{"\\t", []string{"/usr/bin/date", "+\\%T"}},
			}

			for _, variable := range vars {
				if strings.Contains(issue, variable.value) {
					output, _ := exec.Command(variable.command[0], variable.command[1:]...).Output()

					issue = strings.ReplaceAll(issue, variable.value, strings.TrimSpace(string(output)))
				}
			}

			fmt.Print(issue)
		}
	}
}
