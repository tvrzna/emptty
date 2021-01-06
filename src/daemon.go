package src

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const (
	strCleanScreen = "\x1b[H\x1b[2J"
)

// IssueVariable defines list of all escape sequences found in issue file
type issueVariable struct {
	issue string
	char  byte
	arg   string
}

// Starts emptty as daemon spawning emptty on defined TTY.
func startDaemon(conf *config) *os.File {
	fTTY, err := os.OpenFile("/dev/tty"+conf.strTTY(), os.O_RDWR, 0700)
	if err != nil {
		log.Fatal(err)
	}

	clearScreen(fTTY)

	os.Stdout = fTTY
	os.Stderr = fTTY
	os.Stdin = fTTY

	setColors(conf.fgColor, conf.bgColor)
	clearScreen(fTTY)

	fmt.Println()
	if conf.printIssue {
		printIssue()
		setColors(conf.fgColor, conf.bgColor)
	}

	switchTTY(conf)

	return fTTY
}

// Stops daemon mode and closes opened TTY.
func stopDaemon(conf *config, fTTY *os.File) {
	resetColors()
	clearScreen(fTTY)

	if fTTY != nil {
		fTTY.Close()
	}
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
		if err == nil {
			issue := string(bIssue)
			issue = evaluateIssueVars(issue, findUniqueIssueVars(issue))

			if issue[len(issue)-2:] == "\n\n" {
				issue = issue[:len(issue)-1]
			}

			fmt.Print(revertColorEscaping(issue))
		}
	}
}

// Finds all unique issue escape sequences
func findUniqueIssueVars(issue string) []*issueVariable {
	var result []*issueVariable
	var knownIssues []string

	saveData := false
	var buffer strings.Builder
	var varName byte
	var arg strings.Builder

	for i := 0; i < len(issue); i++ {
		b := issue[i]

		if b == '\\' {
			saveData = true
			buffer.Reset()
			arg.Reset()
		}

		if saveData {
			if i > 0 {
				if issue[i-1] == '\\' {
					varName = b
				} else if b != '{' && b != '}' && b != '\\' {
					arg.WriteByte(b)
				}
			}
			buffer.WriteByte(b)
			if (i < len(issue) && i > 0 && issue[i-1] == '\\' && issue[i+1] != '{') || b == '}' {
				if !contains(knownIssues, buffer.String()) {
					result = append(result, &issueVariable{buffer.String(), varName, arg.String()})
					knownIssues = append(knownIssues, buffer.String())
				}

				saveData = false
			}
		}
	}

	return result
}

// Evaluates outputs for all known escape sequences and return replaced issue
func evaluateIssueVars(issue string, issueVars []*issueVariable) string {
	result := issue

	sort.Slice(issueVars, func(i int, j int) bool {
		return len(issueVars[i].arg) > len(issueVars[j].arg)
	})

	for _, issueVar := range issueVars {
		output := ""
		processed := true

		switch issueVar.char {
		case 'd':
			output = runSimpleCmd([]string{"/bin/date"})
		case 'l':
			output = runSimpleCmd([]string{"/bin/ps", "-p", strconv.Itoa(os.Getpid()), "-o", "tty", "--no-headers"})
		case 'm':
			output = runSimpleCmd([]string{"/usr/bin/uname", "-m"})
		case 'n':
			output = runSimpleCmd([]string{"/usr/bin/uname", "-n"})
		case 'r':
			output = runSimpleCmd([]string{"/usr/bin/uname", "-r"})
		case 's':
			output = runSimpleCmd([]string{"/usr/bin/uname", "-s"})
		case 't':
			output = runSimpleCmd([]string{"/bin/date", "+%T"})
		default:
			processed = false
		}

		if processed {
			result = strings.ReplaceAll(result, issueVar.issue, output)
		}
	}

	return result
}
