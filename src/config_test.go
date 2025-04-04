package src

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	loadConfig(getTestingPath("conf"))
	conf := loadConfig(loadConfigPath([]string{"-c", getTestingPath("conf")}))

	if conf.Tty != 14 || conf.strTTY() != "14" {
		t.Error("TestLoadConfig: TTY value is not correct")
	}

	if !conf.SwitchTTY {
		t.Error("TestLoadConfig: SWITCH_TTY value is not correct")
	}

	if !conf.PrintIssue {
		t.Error("TestLoadConfig: PRINT_ISSUE value is not correct")
	}

	if conf.PrintMotd {
		t.Error("TestLoadConfig: PRINT_MOTD value is not correct")
	}

	if conf.DefaultUser != "emptty-user" {
		t.Error("TestLoadConfig: DEFAULT_USER value is not correct")
	}

	if conf.Autologin {
		t.Error("TestLoadConfig: AUTOLOGIN value is not correct")
	}

	if conf.AutologinSession != "none" {
		t.Error("TestLoadConfig: AUTOLOGIN_SESSION value is not correct")
	}

	if conf.AutologinSessionEnv != Undefined {
		t.Error("TestLoadConfig: AUTOLOGIN_SESSION_ENV value is not correct")
	}

	if conf.Lang != "en_US.UTF-8" {
		t.Error("TestLoadConfig: LANG value is not correct")
	}

	if !conf.DbusLaunch {
		t.Error("TestLoadConfig: DBUS_LAUNCH value is not correct")
	}

	if !conf.XinitrcLaunch {
		t.Error("TestLoadConfig: XINITRC_LAUNCH value is not correct")
	}

	if !conf.VerticalSelection {
		t.Error("TestLoadConfig: VERTICAL_SELECTION value is not correct")
	}

	if conf.Logging != Disabled {
		t.Error("TestLoadConfig: LOGGING value is not correct")
	}

	if conf.XorgArgs != "-none" {
		t.Error("TestLoadConfig: XORG_ARGS value is not correct")
	}

	if conf.LoggingFile != "/dev/null" {
		t.Error("TestLoadConfig: LOGGING_FILE value is not correct")
	}

	if !conf.DynamicMotd {
		t.Error("TestLoadConfig: DYNAMIC_MOTD value is not correct")
	}

	if conf.FgColor != "31" {
		t.Error("TestLoadConfig: FG_COLOR value is not correct")
	}

	if conf.BgColor != "44" {
		t.Error("TestLoadConfig: BG_COLOR value is not correct")
	}

	if conf.DisplayStartScript != "/usr/bin/none-start" {
		t.Error("TestLoadConfig: DISPLAY_START_SCRIPT value is not correct")
	}

	if conf.DisplayStopScript != "/usr/bin/none" {
		t.Error("TestLoadConfig: DISPLAY_STOP_SCRIPT value is not correct")
	}

	if !conf.EnableNumlock {
		t.Error("TestLoadConfig: ENABLE_NUMLOCK value is not correct")
	}

	if conf.SessionErrLog != Appending {
		t.Error("TestLoadConfig: SESSION_ERROR_LOGGING value is not correct")
	}

	if conf.SessionErrLogFile != "/dev/null" {
		t.Error("TestLoadConfig: SESSION_ERROR_LOGGING_FILE value is not correct")
	}

	if conf.NoXdgFallback {
		t.Error("TestLoadConfig: NO_XDG_FALLBACK value is not correct")
	}

	if conf.DefaultXauthority {
		t.Error("TestLoadConfig: DEFAULT_XAUTHORITY value is not correct")
	}

	if !conf.RootlessXorg {
		t.Error("TestLoadConfig: ROOTLESS_XORG value is not correct")
	}

	if !conf.IdentifyEnvs {
		t.Error("TestLoadConfig: IDENTIFY_ENVS value is not correct")
	}

	if conf.AutologinMaxRetry != -1 {
		t.Error("TestLoadConfig: AUTOLOGIN_MAX_RETRY value is not correct")
	}

	if conf.MotdPath != "/dev/null/static" {
		t.Error("TestLoadConfig: MOTD_PATH value is not correct")
	}

	if conf.DynamicMotdPath != "/dev/null/dynamic" {
		t.Error("TestLoadConfig: DYNAMIC_MOTD_PATH value is not correct")
	}

	if conf.DefaultSession != "/usr/bin/no-login" {
		t.Error("TestLoadConfig: DEFAULT_SESSION value is not correct")
	}

	if conf.DefaultSessionEnv != Wayland {
		t.Error("TestLoadConfig: DEFAULT_SESSION_ENV value is not correct")
	}

	if !conf.HideEnterLogin {
		t.Error("TestLoadConfig: HIDE_ENTER_LOGIN value is not correct")
	}

	if conf.HideEnterPassword {
		t.Error("TestLoadConfig: HIDE_ENTER_PASSWORD value is not correct")
	}

	if !conf.AlwaysDbusLaunch {
		t.Error("TestLoadConfig: ALWAYS_DBUS_LAUNCH value is not correct")
	}

	if conf.XorgSessionsPath != "/dev/null" {
		t.Error("TestLoadConfig: XORG_SESSIONS_PATH value is not correct")
	}

	if conf.WaylandSessionsPath != "/dev/zero" {
		t.Error("TestLoadConfig: WAYLAND_SESSIONS_PATH value is not correct")
	}

	if conf.SelectLastUser != PerTty {
		t.Error("TestLoadConfig: SELECT_LAST_USER value is not correct")
	}

	if !conf.AutoSelection {
		t.Error("TestLoadConfig: AUTO_SELECTION value is not correct")
	}

	if conf.AutologinRtryPeriod != 2 {
		t.Error("TestLoadConfig: AUTOLOGIN_RETRY_PERIOD value is not correct")
	}
}

func TestLangLoadConfig(t *testing.T) {
	lang := os.Getenv(envLang)
	if lang == "" {
		lang = "C.UTF-8"
	}

	os.Setenv(envLang, "")
	conf := loadConfig(getTestingPath("non-existing-conf"))
	if conf.Lang != "en_US.UTF-8" {
		t.Error("TestLangLoadConfig: fallback language is not correct -", conf.Lang)
	}

	os.Setenv(envLang, lang)
	conf = loadConfig(getTestingPath("non-existing-conf"))
	if conf.Lang != lang {
		t.Error("TestLangLoadConfig: fallback language is not correct -", conf.Lang)
	}
}

func TestParseTTY(t *testing.T) {
	var tty int

	tty = parseTTY("", "6")
	if tty != 6 {
		t.Error("TestParseTTY: wrong default value")
	}

	tty = parseTTY("7", "6")
	if tty != 7 {
		t.Error("TestParseTTY: wrong parsed value")
	}

	tty = parseTTY("aaa", "bbb")
	if tty != 0 {
		t.Error("TestParseTTY: wrong fallback value")
	}
}

func TestTtyPath(t *testing.T) {
	c := &config{Tty: 15}
	if c.ttyPath() != "/dev/tty15" {
		t.Error("TestTtyPath: unexpected result from ttyPath()")
	}
}
