package src

import (
	"os"
	"strconv"
)

const (
	confTTYnumber          = "TTY_NUMBER"
	confSwitchTTY          = "SWITCH_TTY"
	confPrintIssue         = "PRINT_ISSUE"
	confDefaultUser        = "DEFAULT_USER"
	confAutologin          = "AUTOLOGIN"
	confAutologinSession   = "AUTOLOGIN_SESSION"
	confLang               = "LANG"
	confDbusLaunch         = "DBUS_LAUNCH"
	confXinitrcLaunch      = "XINITRC_LAUNCH"
	confVerticalSelection  = "VERTICAL_SELECTION"
	confLogging            = "LOGGING"
	confXorgArgs           = "XORG_ARGS"
	confLoggingFile        = "LOGGING_FILE"
	confDynamicMotd        = "DYNAMIC_MOTD"
	confFgColor            = "FG_COLOR"
	confBgColor            = "BG_COLOR"
	confDisplayStartScript = "DISPLAY_START_SCRIPT"
	confDisplayStopScript  = "DISPLAY_STOP_SCRIPT"

	pathConfigFile = "/etc/emptty/conf"

	constLogDefault   = "default"
	constLogAppending = "appending"
	constLogDisabled  = "disabled"
)

// enLogging defines possible option how to handle configuration.
type enLogging int

const (
	// Default represents saving into new file and backing up older with suffix
	Default enLogging = iota + 1

	// Appending represents saving all logs into same file
	Appending

	// Disabled represents disabled logging
	Disabled
)

// config defines structure of application configuration.
type config struct {
	daemonMode         bool
	defaultUser        string
	autologin          bool
	autologinSession   string
	tty                int
	switchTTY          bool
	printIssue         bool
	lang               string
	dbusLaunch         bool
	xinitrcLaunch      bool
	verticalSelection  bool
	logging            enLogging
	xorgArgs           string
	loggingFile        string
	dynamicMotd        bool
	fgColor            string
	bgColor            string
	displayStartScript string
	displayStopScript  string
}

// LoadConfig handles loading of application configuration.
func loadConfig(path string) *config {
	c := config{
		daemonMode:         false,
		tty:                0,
		switchTTY:          true,
		printIssue:         true,
		defaultUser:        "",
		autologin:          false,
		autologinSession:   "",
		dbusLaunch:         true,
		xinitrcLaunch:      false,
		verticalSelection:  false,
		logging:            Default,
		xorgArgs:           "",
		loggingFile:        "",
		dynamicMotd:        false,
		fgColor:            "",
		bgColor:            "",
		displayStartScript: "",
		displayStopScript:  "",
	}

	defaultLang := os.Getenv(envLang)
	if defaultLang != "" {
		c.lang = defaultLang
	} else {
		c.lang = "en_US.UTF-8"
	}

	if fileExists(path) {
		err := readProperties(path, func(key string, value string) {
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
			case confDynamicMotd:
				c.dynamicMotd = parseBool(value, "false")
			case confFgColor:
				c.fgColor = convertColor(sanitizeValue(value, ""), true)
			case confBgColor:
				c.bgColor = convertColor(sanitizeValue(value, ""), false)
			case confDisplayStartScript:
				c.displayStartScript = sanitizeValue(value, "")
			case confDisplayStopScript:
				c.displayStopScript = sanitizeValue(value, "")
			}
		})
		handleErr(err)
	}

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
