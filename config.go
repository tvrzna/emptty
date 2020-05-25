package main

import (
	"os"
	"strconv"
	"strings"
)

// enEnvironment defines possible Environments.
type enEnvironment int

const (
	// Xorg represents Xorg environment
	Xorg enEnvironment = iota + 1

	// Wayland represents Wayland environment
	Wayland
)

const (
	confTTYnumber   = "TTY_NUMBER"
	confDefaultUser = "DEFAULT_USER"
	confAutologin   = "AUTOLOGIN"
	confLang        = "LANG"
	confDbusLaunch  = "DBUS_LAUNCH"
)

// config defines structure of application configuration.
type config struct {
	defaultUser string
	autologin   bool
	tty         int
	lang        string
	dbusLaunch  bool
}

// LoadConfig handles loading of application configuration.
func loadConfig() *config {
	c := config{tty: 0, defaultUser: "", autologin: false, lang: "en_US.UTF-8", dbusLaunch: true}

	if fileExists("/etc/emptty/conf") {
		err := readProperties("/etc/emptty/conf", func(key string, value string) {
			switch strings.ToUpper(key) {
			case confTTYnumber:
				c.tty = parseTTY(value, "0")
				break
			case confDefaultUser:
				c.defaultUser = parseDefaultUser(value, "")
				break
			case confAutologin:
				c.autologin = parseBool(value, "false")
				break
			case confLang:
				c.lang = sanitizeValue(value, "en_US.UTF-8")
				break
			case confDbusLaunch:
				c.dbusLaunch = parseBool(value, "true")
				break
			}
		})
		handleErr(err)
	}

	os.Unsetenv(confTTYnumber)
	os.Unsetenv(confDefaultUser)
	os.Unsetenv(confAutologin)
	os.Unsetenv(confDbusLaunch)

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

// Parse input env and selects corresponding environment.
func parseEnv(env string, defaultValue string) enEnvironment {
	switch sanitizeValue(env, defaultValue) {
	case "wayland":
		return Wayland
	case "xorg":
		return Xorg
	}
	return Xorg
}

// Stringify enEnvironment value
func stringifyEnv(env enEnvironment) string {
	switch env {
	case Xorg:
		return "xorg"
	case Wayland:
		return "wayland"
	}
	return "xorg"
}

// Parse boolean values
func parseBool(autologin string, defaultValue string) bool {
	val, err := strconv.ParseBool(sanitizeValue(autologin, defaultValue))
	if err != nil {
		return false
	}
	return val
}

// Parse default user.
func parseDefaultUser(defaultUser string, defaultValue string) string {
	return sanitizeValue(defaultUser, defaultValue)
}

// Sanitize value.
func sanitizeValue(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
