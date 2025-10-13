package src

import "strings"

const (
	constEnvXorg    = "xorg"
	constEnvWayland = "wayland"

	constEnvSUndefined  = "Undefined"
	constEnvSXorg       = "Xorg"
	constEnvSWayland    = "Wayland"
	constEnvSCustom     = "Custom"
	constEnvSUserCustom = "User Custom"

	constEnvSTXorg    = "x11"
	constEnvSTWayland = "wayland"
)

// enEnvironment defines possible Environments.
type enEnvironment byte

var defaultEnvValue = Undefined

const (
	// Undefined represents no environment
	Undefined enEnvironment = iota

	// Xorg represents Xorg environment
	Xorg

	// Wayland represents Wayland environment
	Wayland

	// Custom represents custom desktops, only helper before real env is loaded
	Custom

	// UserCustom represents user's desktops, only helper before real env is loaded
	UserCustom
)

// Returns default environment as string value.
func defaultEnv() string {
	return constEnvXorg
}

// Parse input env and selects corresponding environment.
func parseEnv(env, defaultValue string) enEnvironment {
	switch strings.ToLower(sanitizeValue(env, defaultValue)) {
	case constEnvXorg:
		return Xorg
	case constEnvWayland:
		return Wayland
	}
	return defaultEnvValue
}

// Stringify enEnvironment value.
func (e enEnvironment) stringify() string {
	switch e {
	case Xorg:
		return constEnvXorg
	case Wayland:
		return constEnvWayland
	}
	return defaultEnv()
}

// String value of enEnvironment
func (e enEnvironment) string() string {
	strings := []string{constEnvSUndefined, constEnvSXorg, constEnvSWayland, constEnvSCustom, constEnvSUserCustom}
	return strings[e]
}

// Session type of enEnvironment
func (e enEnvironment) sessionType() string {
	strings := []string{"", constEnvSTXorg, constEnvSTWayland, "", ""}
	return strings[e]
}
