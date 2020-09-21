package main

import (
	"os"
	"strconv"
)

const (
	confTTYnumber         = "TTY_NUMBER"
	confSwitchTTY         = "SWITCH_TTY"
	confPrintIssue        = "PRINT_ISSUE"
	confDefaultUser       = "DEFAULT_USER"
	confAutologin         = "AUTOLOGIN"
	confAutologinSession  = "AUTOLOGIN_SESSION"
	confLang              = "LANG"
	confDbusLaunch        = "DBUS_LAUNCH"
	confXinitrcLaunch     = "XINITRC_LAUNCH"
	confVerticalSelection = "VERTICAL_SELECTION"
	confLogging           = "LOGGING"
	confXorgArgs          = "XORG_ARGS"
	confLoggingFile       = "LOGGING_FILE"

	pathConfigFile = "/etc/emptty/conf"

	constLogDefault   = "default"
	constLogAppending = "appending"
	constLogDisabled  = "disabled"
)

// enLogging defines possible option how to handle configuration.
type enLogging int

const (
	Default enLogging = iota + 1
	Appending
	Disabled
)

// config defines structure of application configuration.
type config struct {
	defaultUser       string
	autologin         bool
	autologinSession  string
	tty               int
	switchTTY         bool
	printIssue        bool
	lang              string
	dbusLaunch        bool
	xinitrcLaunch     bool
	verticalSelection bool
	logging           enLogging
	xorgArgs          string
	loggingFile       string
}

// LoadConfig handles loading of application configuration.
func loadConfig() *config {
	c := config{
		tty:              0,
		switchTTY:        true,
		printIssue:       true,
		defaultUser:      "",
		autologin:        false,
		autologinSession: "",
		lang:             "en_US.UTF-8",
		dbusLaunch:       true,
		logging:          Default,
		xorgArgs:         "",
		loggingFile:      "",
	}

	if fileExists(pathConfigFile) {
		err := readProperties(pathConfigFile, func(key string, value string) {
			switch key {
			case confTTYnumber:
				c.tty = parseTTY(value, "0")
			case confSwitchTTY:
				c.switchTTY = parseBool(value, "true")
			case confPrintIssue:
				c.printIssue = parseBool(value, "true")
			case confDefaultUser:
				c.defaultUser = sanitizeValue(value, "")
			case confAutologin:
				c.autologin = parseBool(value, "false")
			case confAutologinSession:
				c.autologinSession = sanitizeValue(value, "")
			case confLang:
				c.lang = sanitizeValue(value, "en_US.UTF-8")
			case confDbusLaunch:
				c.dbusLaunch = parseBool(value, "true")
			case confXinitrcLaunch:
				c.xinitrcLaunch = parseBool(value, "false")
			case confVerticalSelection:
				c.verticalSelection = parseBool(value, "false")
			case confLogging:
				c.logging = parseLogging(value, constLogDefault)
			case confXorgArgs:
				c.xorgArgs = sanitizeValue(value, "")
			case confLoggingFile:
				c.loggingFile = sanitizeValue(value, "")
			}
		})
		handleErr(err)
	}

	os.Unsetenv(confTTYnumber)
	os.Unsetenv(confSwitchTTY)
	os.Unsetenv(confPrintIssue)
	os.Unsetenv(confDefaultUser)
	os.Unsetenv(confAutologin)
	os.Unsetenv(confAutologinSession)
	os.Unsetenv(confDbusLaunch)
	os.Unsetenv(confVerticalSelection)
	os.Unsetenv(confLogging)
	os.Unsetenv(confXorgArgs)
	os.Unsetenv(confLoggingFile)

	return &c
}

// Parse TTY number.
func parseTTY(tty string, defaultValue string) int {
	val, err := strconv.ParseInt(sanitizeValue(tty, defaultValue), 10, 32)
	if err != nil {
		return 0
	}
	return int(val)
}

// Parse boolean values.
func parseBool(strBool string, defaultValue string) bool {
	val, err := strconv.ParseBool(sanitizeValue(strBool, defaultValue))
	if err != nil {
		return false
	}
	return val
}

// Parse logging option
func parseLogging(strLogging string, defaultValue string) enLogging {
	val := sanitizeValue(strLogging, defaultValue)
	switch val {
	case constLogDisabled:
		return Disabled
	case constLogAppending:
		return Appending
	case constLogDefault:
		return Default
	}
	return Default
}

// Returns TTY number converted to string
func (c *config) strTTY() string {
	return strconv.Itoa(c.tty)
}
