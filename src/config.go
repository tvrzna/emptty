package src

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	pathConfigFile = "/etc/emptty/conf"
)

// config defines structure of application configuration.
type config struct {
	DaemonMode          bool
	Autologin           bool             `config:"AUTOLOGIN" parser:"ParseBool" string:"StringBool" default:"false"`
	SwitchTTY           bool             `config:"SWITCH_TTY" parser:"ParseBool" string:"StringBool" default:"true"`
	PrintIssue          bool             `config:"PRINT_ISSUE" parser:"ParseBool" string:"StringBool" default:"true"`
	PrintMotd           bool             `config:"PRINT_MOTD" parser:"ParseBool" string:"StringBool" default:"true"`
	DbusLaunch          bool             `config:"DBUS_LAUNCH" parser:"ParseBool" string:"StringBool" default:"true"`
	AlwaysDbusLaunch    bool             `config:"ALWAYS_DBUS_LAUNCH" parser:"ParseBool" string:"StringBool" default:"false"`
	XinitrcLaunch       bool             `config:"XINITRC_LAUNCH" parser:"ParseBool" string:"StringBool" default:"false"`
	VerticalSelection   bool             `config:"VERTICAL_SELECTION" parser:"ParseBool" string:"StringBool" default:"false"`
	IndentSelection     int              `config:"INDENT_SELECTION" parser:"ParseInt" string:"StringInt" default:"0"`
	DynamicMotd         bool             `config:"DYNAMIC_MOTD" parser:"ParseBool" string:"StringBool" default:"false"`
	EnableNumlock       bool             `config:"ENABLE_NUMLOCK" parser:"ParseBool" string:"StringBool" default:"false"`
	NoXdgFallback       bool             `config:"NO_XDG_FALLBACK" parser:"ParseBool" string:"StringBool" default:"false"`
	DefaultXauthority   bool             `config:"DEFAULT_XAUTHORITY" parser:"ParseBool" string:"StringBool" default:"false"`
	RootlessXorg        bool             `config:"ROOTLESS_XORG" parser:"ParseBool" string:"StringBool" default:"false"`
	IdentifyEnvs        bool             `config:"IDENTIFY_ENVS" parser:"ParseBool" string:"StringBool" default:"false"`
	HideEnterLogin      bool             `config:"HIDE_ENTER_LOGIN" parser:"ParseBool" string:"StringBool" default:"false"`
	HideEnterPassword   bool             `config:"HIDE_ENTER_PASSWORD" parser:"ParseBool" string:"StringBool" default:"false"`
	AutoSelection       bool             `config:"AUTO_SELECTION" parser:"ParseBool" string:"StringBool" default:"false"`
	AllowCommands       bool             `config:"ALLOW_COMMANDS" parser:"ParseBool" string:"StringBool" default:"true"`
	DefaultEnv          enEnvironment    `config:"DEFAULT_ENV" parser:"ParseDefaultEnv" string:"StringEnv" default:"" priority:"true"`
	DefaultSessionEnv   enEnvironment    `config:"DEFAULT_SESSION_ENV" parser:"ParseEnv" string:"StringEnv" default:""`
	AutologinSessionEnv enEnvironment    `config:"AUTOLOGIN_SESSION_ENV" parser:"ParseEnv" string:"StringEnv" default:""`
	Logging             enLogging        `config:"LOGGING" parser:"ParseLogging" string:"StringLog" default:"rotate"`
	SessionErrLog       enLogging        `config:"SESSION_ERROR_LOGGING" parser:"ParseLogging" string:"StringLog" default:"disabled"`
	AutologinMaxRetry   int              `config:"AUTOLOGIN_MAX_RETRY" parser:"ParseInt" string:"StringInt" default:"2"`
	AutologinRtryPeriod int              `config:"AUTOLOGIN_RETRY_PERIOD" parser:"ParsePositiveInt" string:"StringInt" default:"2"`
	Tty                 int              `config:"TTY_NUMBER" parser:"ParseTTY" string:"StringInt" default:"7"`
	DefaultUser         string           `config:"DEFAULT_USER" parser:"SanitizeValue" default:""`
	DefaultSession      string           `config:"DEFAULT_SESSION" parser:"SanitizeValue" default:""`
	AutologinSession    string           `config:"AUTOLOGIN_SESSION" parser:"SanitizeValue" default:""`
	Lang                string           `config:"LANG" parser:"SanitizeValue" default:""`
	LoggingFile         string           `config:"LOGGING_FILE" parser:"SanitizeValue" default:"/var/log/emptty/[TTY_NUMBER].log"`
	XorgArgs            string           `config:"XORG_ARGS" parser:"SanitizeValue" default:""`
	DynamicMotdPath     string           `config:"DYNAMIC_MOTD_PATH" parser:"SanitizeValue" default:"/etc/emptty/motd-gen.sh"`
	MotdPath            string           `config:"MOTD_PATH" parser:"SanitizeValue" default:"/etc/emptty/motd"`
	FgColor             string           `config:"FG_COLOR" parser:"ConvertFgColor" string:"StringFgColor" default:""`
	BgColor             string           `config:"BG_COLOR" parser:"ConvertBgColor" string:"StringBgColor" default:""`
	DisplayStartScript  string           `config:"DISPLAY_START_SCRIPT" parser:"SanitizeValue" default:""`
	DisplayStopScript   string           `config:"DISPLAY_STOP_SCRIPT" parser:"SanitizeValue" default:""`
	SessionErrLogFile   string           `config:"SESSION_ERROR_LOGGING_FILE" parser:"SanitizeValue" default:"/var/log/emptty/session-errors.[TTY_NUMBER].log"`
	XorgSessionsPath    string           `config:"XORG_SESSIONS_PATH" parser:"SanitizeValue" default:"/usr/share/xsessions/"`
	WaylandSessionsPath string           `config:"WAYLAND_SESSIONS_PATH" parser:"SanitizeValue" default:"/usr/share/wayland-sessions/"`
	SelectLastUser      enSelectLastUser `config:"SELECT_LAST_USER" parser:"ParseSelectLastUser" string:"StringLastUser" default:"false"`
	CmdPoweroff         string           `config:"CMD_POWEROFF" parser:"SanitizeValue" default:"poweroff"`
	CmdReboot           string           `config:"CMD_REBOOT" parser:"SanitizeValue" default:"reboot"`
	CmdSuspend          string           `config:"CMD_SUSPEND" parser:"SanitizeValue" default:""`
}

// LoadConfig handles loading of application configuration.
func loadConfig(path string) *config {
	c := config{}

	var configMap map[string]string
	var err error
	if path != "" && fileExists(path) {
		configMap, err = readPropertiesToMap(path)
		if err != nil {
			logFatal(err)
		}
	}

	configType := reflect.TypeOf(c)
	configValue := reflect.ValueOf(&c)

	processFields := func(priority bool) {
		for i := 0; i < configType.NumField(); i++ {
			field := configType.Field(i)

			priorityValue := field.Tag.Get("priority")
			if (priority && priorityValue != "true") || (!priority && priorityValue == "true") {
				continue
			}

			configParam := field.Tag.Get("config")
			parserName := field.Tag.Get("parser")
			defaultValue := field.Tag.Get("default")
			if parserName != "" && configParam != "" {
				settingValue, exists := configMap[configParam]
				if !exists {
					settingValue = defaultValue
				}

				parser := configValue.MethodByName(parserName)
				if parser.Kind() != reflect.Invalid {
					val := parser.Call([]reflect.Value{reflect.ValueOf(settingValue), reflect.ValueOf(defaultValue)})[0]
					configValue.Elem().Field(i).Set(val)
				}
			}
		}
	}

	processFields(true)
	processFields(false)

	if c.Lang == "" {
		defaultLang := os.Getenv(envLang)
		if defaultLang != "" {
			c.Lang = defaultLang
		} else {
			c.Lang = "en_US.UTF-8"
		}
	}

	return &c
}

// Parse TTY number.
func parseTTY(tty, defaultValue string) int {
	val, err := strconv.ParseInt(sanitizeValue(tty, defaultValue), 10, 32)
	if err != nil {
		return 0
	}
	return int(val)
}

// Parses TTY from string to int.
func (c *config) ParseTTY(value, defaultValue string) int {
	return parseTTY(value, defaultValue)
}

// Sanitezes the string value, if value is empty, the defaultValue is returned.
func (c *config) SanitizeValue(value, defaultValue string) string {
	return sanitizeValue(value, defaultValue)
}

// Parses bool value from string.
func (c *config) ParseBool(value, defaultValue string) bool {
	return parseBool(value, defaultValue)
}

// Parses int value from string.
func (c *config) ParseInt(value, defaultValue string) int {
	result, _ := strconv.Atoi(sanitizeValue(value, defaultValue))
	return result
}

// Parses only positive int value from string.
func (c *config) ParsePositiveInt(value, defaultValue string) int {
	result, _ := strconv.Atoi(sanitizeValue(value, defaultValue))
	if result <= 0 {
		result, _ = strconv.Atoi(defaultValue)
	}
	return result
}

// Parses logging type from string.
func (c *config) ParseLogging(value, defaultValue string) enLogging {
	return parseLogging(value, defaultValue)
}

// Parses environment type from string
func (c *config) ParseEnv(value, defaultValue string) enEnvironment {
	if value == "" {
		return Undefined
	}
	return parseEnv(value, defaultValue)
}

// Parses default environment type from string
func (c *config) ParseDefaultEnv(value, defaultValue string) enEnvironment {
	switch strings.ToLower(sanitizeValue(value, defaultValue)) {
	case constEnvXorg:
		defaultEnvValue = Xorg
	case constEnvWayland:
		defaultEnvValue = Wayland
	default:
		// The default of default, could be once defined by build tag.
		defaultEnvValue = Xorg
	}
	return defaultEnvValue
}

// Coverts string foreground color name into ANSI color value.
func (c *config) ConvertFgColor(value, defaultValue string) string {
	return convertColor(sanitizeValue(value, defaultValue), true)
}

// Converts string background color name into ANSI color value.
func (c *config) ConvertBgColor(value, defaultValue string) string {
	return convertColor(sanitizeValue(value, defaultValue), false)
}

// Returns TTY number converted to string
func (c *config) strTTY() string {
	return strconv.Itoa(c.Tty)
}

// Returns path to TTY
func (c *config) ttyPath() string {
	return "/dev/tty" + c.strTTY()
}

// Parses select last user config option.
func (c *config) ParseSelectLastUser(value, defaultValue string) enSelectLastUser {
	switch strings.ToLower(sanitizeValue(value, defaultValue)) {
	case constEnSelectLastUserPerTTy:
		return PerTty
	case constEnSelectLastUserGlobal:
		return Global
	}
	return False
}

func (c *config) printConfig() {
	configType := reflect.TypeOf(*c)
	configValue := reflect.ValueOf(*c)
	confValue := reflect.ValueOf(c)

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		param := field.Tag.Get("config")
		if param == "" {
			continue
		}

		value := configValue.Field(i).Interface()
		stringName := field.Tag.Get("string")
		if stringName != "" {
			parser := confValue.MethodByName(stringName)
			if parser.Kind() != reflect.Invalid {
				value = parser.Call([]reflect.Value{reflect.ValueOf(value)})[0]
			}
		}
		fmt.Printf("%s=%s\n", param, value)
	}
}

func (c *config) StringBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (c *config) StringEnv(value enEnvironment) string {
	return value.stringify()
}

func (c *config) StringLog(value enLogging) string {
	return value.stringify()
}

func (c *config) StringLastUser(value enSelectLastUser) string {
	return []string{constEnSelectLastUserFalse, constEnSelectLastUserPerTTy, constEnSelectLastUserGlobal}[int(value)]
}

func (c *config) StringInt(value int) string {
	return strconv.Itoa(value)
}

func (c *config) StringFgColor(value string) string {
	return stringColor(value, true)
}

func (c *config) StringBgColor(value string) string {
	return stringColor(value, false)
}

//TODO: escaping
