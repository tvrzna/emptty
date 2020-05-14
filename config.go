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
	envTTYnumber   = "TTY_NUMBER"
	envDefaultUser = "DEFAULT_USER"
	envAutologin   = "AUTOLOGIN"
	envEnvironment = "ENVIRONMENT"
)

// config defines structure of application configuration.
type config struct {
	environment enEnvironment
	defaultUser string
	autologin   bool
	tty         int
}

// LoadConfig handles loading of application configuration.
func loadConfig() *config {
	c := config{}

	c.environment = parseEnv(os.Getenv(envEnvironment))
	c.tty = parseTTY(os.Getenv(envTTYnumber))
	c.defaultUser = parseDefaultUser(os.Getenv(envDefaultUser))
	autologin := parseAutologin(os.Getenv(envAutologin))
	if autologin && c.defaultUser != "" {
		c.autologin = autologin
	}

	os.Unsetenv(envEnvironment)
	os.Unsetenv(envTTYnumber)
	os.Unsetenv(envDefaultUser)
	os.Unsetenv(envAutologin)

	return &c
}

// Parse TTY number.
func parseTTY(tty string) int {
	val, err := strconv.ParseInt(sanitizeValue(tty, "0"), 10, 32)
	if err != nil {
		return 0
	}
	return int(val)
}

// Parse input env and selects corresponding environment.
func parseEnv(env string) enEnvironment {
	switch sanitizeValue(env, "xorg") {
	case "wayland":
		return Wayland
	case "xorg":
		return Xorg
	}
	return Xorg
}

// Parse, if autologin is enabled.
func parseAutologin(autologin string) bool {
	val, err := strconv.ParseBool(sanitizeValue(autologin, "false"))
	if err != nil {
		return false
	}
	return val
}

// Parse default user.
func parseDefaultUser(defaultUser string) string {
	return sanitizeValue(defaultUser, "")
}

// Sanitize value.
func sanitizeValue(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}
