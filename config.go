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

	Unknown
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

	tmpConfig := parseConfigFromFile()

	c.environment = parseEnv(os.Getenv(envEnvironment), "unknown")
	if c.environment == Unknown {
		c.environment = tmpConfig.environment
	}

	c.tty = parseTTY(os.Getenv(envTTYnumber), "-1")
	if c.tty == -1 {
		c.tty = tmpConfig.tty
	}

	c.defaultUser = parseDefaultUser(os.Getenv(envDefaultUser), "@@@@")
	if c.defaultUser == "@@@@" {
		c.defaultUser = tmpConfig.defaultUser
	}

	c.autologin = parseAutologin(os.Getenv(envAutologin), "nil")
	if os.Getenv(envAutologin) == "" {
		c.autologin = tmpConfig.autologin
	}

	if c.autologin && c.defaultUser != "" {
		c.autologin = true
	} else {
		c.autologin = false
	}

	os.Unsetenv(envEnvironment)
	os.Unsetenv(envTTYnumber)
	os.Unsetenv(envDefaultUser)
	os.Unsetenv(envAutologin)

	return &c
}

func parseConfigFromFile() *config {
	c := config{environment: Xorg, tty: 0, defaultUser: "", autologin: false}

	err := readProperties("/etc/emptty/conf", func(key string, value string) error {
		switch key {
		case envTTYnumber:
			c.tty = parseTTY(value, "0")
			break
		case envDefaultUser:
			c.defaultUser = parseDefaultUser(value, "")
			break
		case envAutologin:
			c.autologin = parseAutologin(value, "false")
			break
		case envEnvironment:
			c.environment = parseEnv(value, "xorg")
			break
		}
		return nil
	})
	handleErr(err)

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
	return Unknown
}

// Parse, if autologin is enabled.
func parseAutologin(autologin string, defaultValue string) bool {
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
