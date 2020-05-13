package main

import (
	"flag"
)

// EnEnvironment defines possible Environments
type EnEnvironment int
const (
	// Xorg represents Xorg environment
	Xorg EnEnvironment = iota + 1

	// Wayland represents Wayland environment
	Wayland
)

// Config defines structure of application configuration.
type Config struct {
	environment EnEnvironment
	defaultUser string
	autologin bool
	tty int
}

// LoadConfig handles loading of application configuration.
func LoadConfig() *Config {
	c := Config{}

	tty := flag.Int("t", 7, "used TTY")
	env := flag.String("e", "xorg", "environment to be started")
	usr := flag.String("u", "", "preselected user")
	autologin := flag.Bool("a", false, "autologin to user, expects to preselect user")

	flag.Parse()

	c.environment = parseEnv(*env)
	c.defaultUser = *usr
	c.tty = *tty;
	if *autologin && c.defaultUser != "" {
		c.autologin = *autologin
	}

	return &c
}

// Parse input env and selects corresponding environment.
func parseEnv(env string) EnEnvironment {
	switch env {
	case "wayland":
		return Wayland
	case "xorg":
	default:
		return Xorg
	}
	return Xorg
}