package src

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

const (
	strCleanScreen = "\x1b[H\x1b[2J"
	pathIssue      = "/etc/issue"
)

// IssueVariable defines list of all escape sequences found in issue file
type issueVariable struct {
	issue string
	char  byte
	arg   string
}

// Starts emptty as daemon spawning emptty on defined TTY, if allowed.
func startDaemon(conf *config) *os.File {
	if !conf.DaemonMode {
		return nil
	}

	fTTY, err := os.OpenFile(conf.ttyPath(), os.O_RDWR, 0700)
	if err != nil {
		logFatal(err)
	}

	if conf.EnableNumlock {
		setKeyboardLeds(fTTY, false, true, false)
	}

	clearScreen(fTTY)

	os.Stdout = fTTY
	os.Stderr = fTTY
	os.Stdin = fTTY

	setColors(conf.FgColor, conf.BgColor)
	clearScreen(fTTY)

	if conf.PrintIssue {
		fmt.Println()
		printIssue(pathIssue, conf.strTTY())
		setColors(conf.FgColor, conf.BgColor)
	}

	switchTTY(conf)

	return fTTY
}

// Stops daemon mode and closes opened TTY, if allowed
func stopDaemon(conf *config, fTTY *os.File) {
	if !conf.DaemonMode {
		return
	}

	resetColors()
	clearScreen(fTTY)

	if fTTY != nil {
		fTTY.Close()
	}
}

// Clears terminal screen
func clearScreen(w io.Writer) {
	if w == nil {
		fmt.Print(strCleanScreen)
	} else {
		w.Write([]byte(strCleanScreen))
	}
}

// Perform switch to defined TTY, if switchTTY is true and tty is greater than 0.
func switchTTY(conf *config) bool {
	if conf.SwitchTTY && conf.Tty > 0 {
		return chvt(conf.Tty)
	}
	return false
}

// Prints getty issue
func printIssue(path, strTTY string) {
	if fileExists(path) {
		bIssue, err := ioutil.ReadFile(path)
		if err == nil {
			issue := string(bIssue)
			issue = evaluateIssueVars(issue, findUniqueIssueVars(issue), strTTY)

			for issue[len(issue)-2:] == "\n\n" {
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

	for i := 0; i < len(issue); i++ {
		var issueVar *issueVariable
		issueVar, i = findIssueVar(issue, i)
		if issueVar != nil && !contains(knownIssues, issueVar.issue) {
			result = append(result, issueVar)
			knownIssues = append(knownIssues, issueVar.issue)
		}
	}

	return result
}

// Finds single issue escape sequence
func findIssueVar(issue string, i int) (*issueVariable, int) {
	saveData := false
	var j int
	var buffer strings.Builder
	var varName byte
	var arg strings.Builder

	for j = i; j < len(issue); j++ {
		b := issue[j]

		if b == '\\' {
			saveData = true
			buffer.Reset()
			arg.Reset()
		}

		if saveData {
			if j > 0 {
				if issue[j-1] == '\\' {
					varName = b
				} else if b != '{' && b != '}' && b != '\\' {
					arg.WriteByte(b)
				}
			}
			buffer.WriteByte(b)
			if j == (len(issue)-1) || (j < len(issue) && j > 0 && issue[j-1] == '\\' && issue[j+1] != '{') || b == '}' {
				return &issueVariable{buffer.String(), varName, arg.String()}, j
			}
		}
	}
	return nil, j
}

// Evaluates outputs for all known escape sequences and return replaced issue
func evaluateIssueVars(issue string, issueVars []*issueVariable, strTTY string) string {
	result := issue

	sort.Slice(issueVars, func(i int, j int) bool {
		return len(issueVars[i].arg) > len(issueVars[j].arg)
	})

	for _, issueVar := range issueVars {
		if output, processed := evaluateIssueVar(issueVar, strTTY); processed {
			result = strings.ReplaceAll(result, issueVar.issue, output)
		}
	}

	return result
}

// Evaluate single issue variable and return its result value
func evaluateIssueVar(issueVar *issueVariable, strTTY string) (output string, processed bool) {
	output = ""
	processed = true

	switch issueVar.char {
	case 'd':
		output = runSimpleCmd("date")
	case 'l':
		output = getCurrentTTYName(strTTY, false)
	case 'm':
		output = runSimpleCmd("uname", "-m")
	case 'n':
		output = runSimpleCmd("uname", "-n")
	case 'O':
		output = getDnsDomainName()
	case 'r':
		output = runSimpleCmd("uname", "-r")
	case 's':
		output = runSimpleCmd("uname", "-s")
	case 'S':
		output = getOsReleaseValue(issueVar.arg)
	case 't':
		output = runSimpleCmd("date", "+%T")
	case '4', '6':
		output = getIpAddress(issueVar.arg, issueVar.char)
	default:
		processed = false
	}

	return
}
