package src

import (
	"os"
	"strconv"
)

const (
	confTTYnumber          = "TTY_NUMBER"
	confSwitchTTY          = "SWITCH_TTY"
	confPrintIssue         = "PRINT_ISSUE"
	confPrintMotd          = "PRINT_MOTD"
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
	confEnableNumlock      = "ENABLE_NUMLOCK"
	confSessionErrLog      = "SESSION_ERROR_LOGGING"
	confSessionErrLogFile  = "SESSION_ERROR_LOGGING_FILE"
	confNoXdgFallback      = "NO_XDG_FALLBACK"
	confDefaultXauthority  = "DEFAULT_XAUTHORITY"

	pathConfigFile = "/etc/emptty/conf"
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
	printMotd          bool
	lang               string
	dbusLaunch         bool
	xinitrcLaunch      bool
	verticalSelection  bool
	logging            enLogging
	loggingFile        string
	xorgArgs           string
	dynamicMotd        bool
	fgColor            string
	bgColor            string
	displayStartScript string
	displayStopScript  string
	enableNumlock      bool
	sessionErrLog      enLogging
	sessionErrLogFile  string
	noXdgFallback      bool
	defaultXauthority  bool
}

// LoadConfig handles loading of application configuration.
func loadConfig(path string) *config {
	c := config{
		daemonMode:         false,
		tty:                0,
		switchTTY:          true,
		printIssue:         true,
		printMotd:          true,
		defaultUser:        "",
		autologin:          false,
		autologinSession:   "",
		dbusLaunch:         true,
		xinitrcLaunch:      false,
		verticalSelection:  false,
		logging:            Default,
		loggingFile:        "",
		xorgArgs:           "",
		dynamicMotd:        false,
		fgColor:            "",
		bgColor:            "",
		displayStartScript: "",
		displayStopScript:  "",
		enableNumlock:      false,
		sessionErrLog:      Disabled,
		sessionErrLogFile:  "",
		noXdgFallback:      false,
		defaultXauthority:  false,
	}

	defaultLang := os.Getenv(envLang)
	if defaultLang != "" {
		c.lang = defaultLang
	} else {
		c.lang = "en_US.UTF-8"
	}

	if path != "" && fileExists(path) {
		err := readProperties(path, func(key string, value string) {
			switch key {
			case confTTYnumber:
				c.tty = parseTTY(value, "0")
			case confSwitchTTY:
				c.switchTTY = parseBool(value, "true")
			case confPrintIssue:
				c.printIssue = parseBool(value, "true")
			case confPrintMotd:
				c.printMotd = parseBool(value, "true")
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
			case confLoggingFile:
				c.loggingFile = sanitizeValue(value, "")
			case confXorgArgs:
				c.xorgArgs = sanitizeValue(value, "")
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
			case confEnableNumlock:
				c.enableNumlock = parseBool(value, "false")
			case confSessionErrLog:
				c.sessionErrLog = parseLogging(value, constLogDisabled)
			case confSessionErrLogFile:
				c.sessionErrLogFile = sanitizeValue(value, "")
			case confNoXdgFallback:
				c.noXdgFallback = parseBool(value, "false")
			case confDefaultXauthority:
				c.defaultXauthority = parseBool(value, "false")
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

// Returns TTY number converted to string
func (c *config) strTTY() string {
	return strconv.Itoa(c.tty)
}
