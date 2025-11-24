package src

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	pathLastLoggedInUser = "/var/cache/emptty/lastuser"

	constEnSelectLastUserFalse  = "false"
	constEnSelectLastUserPerTTy = "per-tty"
	constEnSelectLastUserGlobal = "global"
)

type enSelectLastUser byte

const (
	// Do not preselect any last user
	False enSelectLastUser = iota

	// Preselect last successfully logged in user per tty
	PerTty

	// Preselect last successfully logged in user per system
	Global
)

type authBase struct {
	command string
}

func (a *authBase) getCommand() string {
	return a.command
}

// Performs input selection user. If saving last user is enabled (PerTty/Global), user is read from defined path and used as predefined value.
func (a *authBase) selectUser(c *config) (string, error) {
	indent := ""
	if c.VerticalSelection && c.IndentSelection > 0 {
		indent = strings.Repeat(" ", c.IndentSelection)
	}

	if c.DefaultUser != "" {
		if !c.HideEnterLogin {
			hostname, _ := os.Hostname()
			fmt.Printf("%s%s login: %s\n", indent, hostname, c.DefaultUser)
		}
		return c.DefaultUser, nil
	}

	lastUser := a.getLastSelectedUser(c)
	if !c.HideEnterLogin {
		hostname, _ := os.Hostname()
		lastUserDisplay := ""
		if lastUser != "" {
			lastUserDisplay = " [" + lastUser + "]"
		}
		fmt.Printf("%s%s login%s: ", indent, hostname, lastUserDisplay)
	}
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	username := input[:len(input)-1]

	if c.AllowCommands && strings.HasPrefix(strings.ReplaceAll(username, "\x1b", ""), ":") {
		a.command = strings.ReplaceAll(username, "\x1b", "")[1:]
		return "", nil
	}

	if lastUser != "" && username == "" {
		username = lastUser
	}
	return username, nil
}

// Gets last selected user with respect to configuration.
func (a *authBase) getLastSelectedUser(c *config) string {
	switch c.SelectLastUser {
	case PerTty:
		return a.readLastUser(pathLastLoggedInUser + "-" + c.strTTY())
	case Global:
		return a.readLastUser(pathLastLoggedInUser)
	}
	return ""
}

// Reads last user from file on path.
func (a *authBase) readLastUser(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			logPrint(err)
		}
		return ""
	}
	lastUser := strings.TrimSpace(string(b))
	if strings.Contains(lastUser, "\n") {
		return ""
	}
	return lastUser
}

// Saves last selected user with respect to configuration.
func (a *authBase) saveLastSelectedUser(c *config, username string) {
	if c.SelectLastUser == False {
		return
	}

	path := pathLastLoggedInUser
	if c.SelectLastUser == PerTty {
		path += "-" + c.strTTY()
	}

	if err := mkDirsForFile(path, 0700); err != nil {
		logPrint(err)
	}
	if err := os.WriteFile(path, []byte(username), 0600); err != nil {
		logPrint(err)
	}
}
