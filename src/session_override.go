package src

import (
	"path/filepath"
	"reflect"
	"strings"
)

const pathSessionOverridesDir = "/etc/emptty/"

// Sanitizes session name to be safely usable as a directory name
func sanitizeSessionName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Returns path to session-specific override dir
func sessionOverrideDir(sessionName string) string {
	sname := sanitizeSessionName(sessionName)
	if sname == "" {
		return ""
	}
	return pathSessionOverridesDir + sname + "/"
}

// Resolves the effective name to use for override lookups, based on
// the .desktop file's basename (without extension), not the display Name=.
// This distinguishes e.g. icewm.desktop from icewm-session.desktop, even
// though both may have Name=IceWM
func sessionNameOf(d *desktop) string {
	if d == nil {
		return ""
	}

	path := d.path
	if path == "" && d.child != nil {
		path = d.child.path
	}
	if path == "" {
		return ""
	}

	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// Applies overrides from a parsed key-value map onto an existing config,
// only touching fields that are actually present in the map.
// Reuses the same "config" struct tags/parsers as the main config loader
func applySessionConfigOverride(c *config, configMap map[string]string) {
	configType := reflect.TypeOf(*c)
	configValue := reflect.ValueOf(c)

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		configParam := field.Tag.Get("config")
		if configParam == "" {
			continue
		}

		settingValue, exists := configMap[configParam]
		if !exists {
			continue
		}

		parserName := field.Tag.Get("parser")
		if parserName != "" {
			parser := configValue.MethodByName(parserName)
			if parser.Kind() != reflect.Invalid {
				val := parser.Call([]reflect.Value{reflect.ValueOf(settingValue), reflect.ValueOf(settingValue)})[0]
				configValue.Elem().Field(i).Set(val)
			}
		} else {
			switch configValue.Elem().Field(i).Type().Kind() {
			case reflect.String:
				configValue.Elem().Field(i).SetString(settingValue)
			case reflect.Bool:
				configValue.Elem().Field(i).SetBool(c.ParseBool(settingValue, settingValue))
			}
		}
	}
}

// If present, loads /etc/emptty/<session>/conf and overrides matching fields on conf
func loadSessionConfOverride(conf *config, sessionName string) {
	dir := sessionOverrideDir(sessionName)
	if dir == "" {
		return
	}

	path := dir + "conf"
	if !fileExists(path) {
		return
	}

	configMap, err := readPropertiesToMap(path)
	if err != nil {
		logPrint(err)
		return
	}

	applySessionConfigOverride(conf, configMap)
	logPrint("Applied session config override from " + path)
}

// if present, loads /etc/emptty/<session>/env and sets each entry as env var for user
func loadSessionEnvOverride(usr *sysuser, sessionName string) {
	dir := sessionOverrideDir(sessionName)
	if dir == "" {
		return
	}

	path := dir + "env"
	if !fileExists(path) {
		return
	}

	err := readProperties(path, func(key, value string) {
		usr.setenv(key, value)
	})
	if err != nil {
		logPrint(err)
		return
	}

	logPrint("Applied session env override from " + path)
}