package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	strCleanScreen = "\x1b[H\x1b[2J"
)

// IssueVariable defines commands being set during printing of issue file
type issueVariable struct {
	value   string
	command []string
}

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

	fmt.Println()
	if conf.printIssue {
		printIssue()
	}

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

			fmt.Print(revertColorEscaping(issue))
			// Clear to default
			fmt.Print("\x1b[0m\n")
		}
	}
}
