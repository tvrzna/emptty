package src

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	lang := os.Getenv(envLang)
	os.Setenv(envLang, "")
	conf := loadConfig(getTestingPath("conf"))
	os.Setenv(envLang, lang)

	if conf.lang != "en_US.UTF-8" {
		t.Error("TestLoadConfig: fallback language is not correct")
	}

	conf = loadConfig(loadConfigPath([]string{"-c", getTestingPath("conf")}))

	if conf.tty != 14 || conf.strTTY() != "14" {
		t.Error("TestLoadConfig: TTY value is not correct")
	}

	if !conf.switchTTY {
		t.Error("TestLoadConfig: SWITCH_TTY value is not correct")
	}

	if !conf.printIssue {
		t.Error("TestLoadConfig: PRINT_ISSUE value is not correct")
	}

	if conf.defaultUser != "emptty-user" {
		t.Error("TestLoadConfig: DEFAULT_USER value is not correct")
	}

	if conf.autologin {
		t.Error("TestLoadConfig: AUTOLOGIN value is not correct")
	}

	if conf.autologinSession != "none" {
		t.Error("TestLoadConfig: AUTOLOGIN_SESSION value is not correct")
	}

	if conf.lang != "en_US.UTF-8" {
		t.Error("TestLoadConfig: LANG value is not correct")
	}

	if !conf.dbusLaunch {
		t.Error("TestLoadConfig: DBUS_LAUNCH value is not correct")
	}

	if !conf.xinitrcLaunch {
		t.Error("TestLoadConfig: XINITRC_LAUNCH value is not correct")
	}

	if !conf.verticalSelection {
		t.Error("TestLoadConfig: VERTICAL_SELECTION value is not correct")
	}

	if conf.logging != Disabled {
		t.Error("TestLoadConfig: LOGGING value is not correct")
	}

	if conf.xorgArgs != "-none" {
		t.Error("TestLoadConfig: XORG_ARGS value is not correct")
	}

	if conf.loggingFile != "/dev/null" {
		t.Error("TestLoadConfig: LOGGING_FILE value is not correct")
	}

	if !conf.dynamicMotd {
		t.Error("TestLoadConfig: DYNAMIC_MOTD value is not correct")
	}

	if conf.fgColor != "31" {
		t.Error("TestLoadConfig: FG_COLOR value is not correct")
	}

	if conf.bgColor != "44" {
		t.Error("TestLoadConfig: BG_COLOR value is not correct")
	}

	if conf.displayStartScript != "/usr/bin/none-start" {
		t.Error("TestLoadConfig: DISPLAY_START_SCRIPT value is not correct")
	}

	if conf.displayStopScript != "/usr/bin/none" {
		t.Error("TestLoadConfig: DISPLAY_STOP_SCRIPT value is not correct")
	}

	if !conf.enableNumlock {
		t.Error("TestLoadConfig: ENABLE_NUMLOCK value is not correct")
	}

	if conf.sessionErrLog != Appending {
		t.Error("TestLoadconfig: SESSION_ERROR_LOGGING value is not correct")
	}

	if conf.sessionErrLogFile != "/dev/null" {
		t.Error("TestLoadconfig: SESSION_ERROR_LOGGING_FILE value is not correct")
	}

	if conf.noXdgFallback {
		t.Error("TestLoadconfig: NO_XDG_FALLBACK value is not correct")
	}

	if conf.defaultXauthority {
		t.Error("TestLoadconfig: DEFAULT_XAUTHORITY value is not correct")
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
