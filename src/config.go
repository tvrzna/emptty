package src

import (
	"os"
	"reflect"
	"strconv"
)

const (
	pathConfigFile = "/etc/emptty/conf"
)

// config defines structure of application configuration.
type config struct {
	DaemonMode          bool
	Autologin           bool          `config:"AUTOLOGIN" parser:"ParseBool" default:"false"`
	SwitchTTY           bool          `config:"SWITCH_TTY" parser:"ParseBool" default:"true"`
	PrintIssue          bool          `config:"PRINT_ISSUE" parser:"ParseBool" default:"true"`
	PrintMotd           bool          `config:"PRINT_MOTD" parser:"ParseBool" default:"true"`
	DbusLaunch          bool          `config:"DBUS_LAUNCH" parser:"ParseBool" default:"true"`
	AlwaysDbusLaunch    bool          `config:"ALWAYS_DBUS_LAUNCH" parser:"ParseBool" default:"false"`
	XinitrcLaunch       bool          `config:"XINITRC_LAUNCH" parser:"ParseBool" default:"false"`
	VerticalSelection   bool          `config:"VERTICAL_SELECTION" parser:"ParseBool" default:"false"`
	DynamicMotd         bool          `config:"DYNAMIC_MOTD" parser:"ParseBool" default:"false"`
	EnableNumlock       bool          `config:"ENABLE_NUMLOCK" parser:"ParseBool" default:"false"`
	NoXdgFallback       bool          `config:"NO_XDG_FALLBACK" parser:"ParseBool" default:"false"`
	DefaultXauthority   bool          `config:"DEFAULT_XAUTHORITY" parser:"ParseBool" default:"false"`
	RootlessXorg        bool          `config:"ROOTLESS_XORG" parser:"ParseBool" default:"false"`
	IdentifyEnvs        bool          `config:"IDENTIFY_ENVS" parser:"ParseBool" default:"false"`
	HideEnterLogin      bool          `config:"HIDE_ENTER_LOGIN" parser:"ParseBool" default:"false"`
	HideEnterPassword   bool          `config:"HIDE_ENTER_PASSWORD" parser:"ParseBool" default:"false"`
	DefaultSessionEnv   enEnvironment `config:"DEFAULT_SESSION_ENV" parser:"ParseEnv" default:""`
	AutologinSessionEnv enEnvironment `config:"AUTOLOGIN_SESSION_ENV" parser:"ParseEnv" default:""`
	Logging             enLogging     `config:"LOGGING" parser:"ParseLogging" default:"default"`
	SessionErrLog       enLogging     `config:"SESSION_ERROR_LOGGING" parser:"ParseLogging" default:"disabled"`
	AutologinMaxRetry   int           `config:"AUTOLOGIN_MAX_RETRY" parser:"ParseInt" default:"2"`
	Tty                 int           `config:"TTY_NUMBER" parser:"ParseTTY" default:"0"`
	DefaultUser         string        `config:"DEFAULT_USER" parser:"SanitizeValue" default:""`
	DefaultSession      string        `config:"DEFAULT_SESSION" parser:"SanitizeValue" default:""`
	AutologinSession    string        `config:"AUTOLOGIN_SESSION" parser:"SanitizeValue" default:""`
	Lang                string        `config:"LANG" parser:"SanitizeValue" default:""`
	LoggingFile         string        `config:"LOGGING_FILE" parser:"SanitizeValue" default:"/var/log/emptty/[TTY_NUMBER].log"`
	XorgArgs            string        `config:"XORG_ARGS" parser:"SanitizeValue" default:""`
	DynamicMotdPath     string        `config:"DYNAMIC_MOTD_PATH" parser:"SanitizeValue" default:"/etc/emptty/motd-gen.sh"`
	MotdPath            string        `config:"MOTD_PATH" parser:"SanitizeValue" default:"/etc/emptty/motd"`
	FgColor             string        `config:"FG_COLOR" parser:"ConvertFgColor" default:""`
	BgColor             string        `config:"BG_COLOR" parser:"ConvertBgColor" default:""`
	DisplayStartScript  string        `config:"DISPLAY_START_SCRIPT" parser:"SanitizeValue" default:""`
	DisplayStopScript   string        `config:"DISPLAY_STOP_SCRIPT" parser:"SanitizeValue" default:""`
	SessionErrLogFile   string        `config:"SESSION_ERROR_LOGGING_FILE" parser:"SanitizeValue" default:"/var/log/emptty/session-errors.[TTY_NUMBER].log"`
	XorgSessionsPath    string        `config:"XORG_SESSIONS_PATH" parser:"SanitizeValue" default:"/usr/share/xsessions/"`
	WaylandSessionsPath string        `config:"WAYLAND_SESSIONS_PATH" parser:"SanitizeValue" default:"/usr/share/wayland-sessions/"`
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

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)

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
