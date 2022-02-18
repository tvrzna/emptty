package src

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	confCommand   = "COMMAND"
	confSelection = "SELECTION"

	desktopExec        = "EXEC"
	desktopName        = "NAME"
	desktopEnvironment = "ENVIRONMENT"
	desktopLang        = "LANG"

	constEnvXorg    = "xorg"
	constEnvWayland = "wayland"

	constEnvSUndefined  = "Undefined"
	constEnvSXorg       = "Xorg"
	constEnvSWayland    = "Wayland"
	constEnvSCustom     = "Custom"
	constEnvSUserCustom = "User Custom"

	pathLastSession       = "/.cache/emptty/last-session"
	pathXorgSessions      = "/usr/share/xsessions/"
	pathWaylandSessions   = "/usr/share/wayland-sessions/"
	pathCustomSessions    = "/etc/emptty/custom-sessions/"
	pathUserCustomSession = "/.config/emptty-custom-sessions/"
)

// enEnvironment defines possible Environments.
type enEnvironment int

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

// desktop defines structure for display environments and window managers.
type desktop struct {
	name      string
	exec      string
	env       enEnvironment
	envOrigin enEnvironment
	isUser    bool
	path      string
	selection bool
	child     *desktop
}

// Gets exec path from desktop and returns true, if command allows dbus-launch.
func (d *desktop) getStrExec() (string, bool) {
	if d.exec != "" {
		return d.exec, true
	}
	return d.path, false
}

// lastSession defines structure for last used session on user login.
type lastSession struct {
	exec string
	env  enEnvironment
}

// Allows to select desktop, which could be selected.
func selectDesktop(usr *sysuser, conf *config) *desktop {
	desktops := listAllDesktops(usr, pathXorgSessions, pathWaylandSessions)
	if len(desktops) == 0 {
		handleStrErr("Not found any installed desktop.")
	}

	lastDesktop := getLastDesktop(usr, desktops)

	if conf.Autologin && conf.AutologinSession != "" {
		d := findAutoselectDesktop(conf.AutologinSession, desktops)
		if d != nil {
			if isLastDesktopForSave(usr, desktops[lastDesktop], d) {
				setUserLastSession(usr, d)
			}
			return d
		}
	}

	for true {
		fmt.Printf("\n")
		printDesktops(conf, desktops)
		fmt.Printf("\nSelect [%d]: ", lastDesktop)

		selection, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		selection = strings.TrimSpace(selection)
		if selection == "" {
			selection = strconv.Itoa(lastDesktop)
		}

		id, err := strconv.ParseUint(selection, 10, 32)
		if err != nil {
			continue
		}
		if int(id) < len(desktops) {
			d := desktops[id]
			if isLastDesktopForSave(usr, desktops[lastDesktop], d) {
				setUserLastSession(usr, d)
			}
			return d
		}
	}
	return nil
}

// Prints list of desktops on screen
func printDesktops(conf *config, desktops []*desktop) {
	dSeparator := ", "
	eSeparator := " "
	if conf.VerticalSelection {
		dSeparator = "\n"
		eSeparator = "\n"
	}

	lastEnv := Undefined
	for i, v := range desktops {
		printSeparator := true
		if conf.IdentifyEnvs && v.envOrigin != lastEnv {
			if i > 0 {
				fmt.Print(eSeparator)
				fmt.Print(eSeparator)
			}
			lastEnv = v.envOrigin
			fmt.Printf("|%s|%s", lastEnv.string(), eSeparator)
			printSeparator = false
		}

		if printSeparator && i > 0 {
			fmt.Print(dSeparator)
		}
		fmt.Printf("[%d] %s", i, v.name)
	}
}

// Finds defined autologinSession in array of desktops by its exec or its name
func findAutoselectDesktop(autologinSession string, desktops []*desktop) *desktop {
	for _, d := range desktops {
		if strings.HasSuffix(d.exec, autologinSession) || autologinSession == d.name {
			return d
		}
	}
	return nil
}

// List all installed desktops and return their exec commands.
func listAllDesktops(usr *sysuser, pathXorgDesktops, pathWaylandDesktops string) []*desktop {
	var result []*desktop

	// load Xorg desktops
	xorgDesktops := listDesktops(pathXorgDesktops, Xorg)
	if xorgDesktops != nil && len(xorgDesktops) > 0 {
		result = append(result, xorgDesktops...)
	}

	// load Wayland desktops
	waylandDesktops := listDesktops(pathWaylandDesktops, Wayland)
	if waylandDesktops != nil && len(waylandDesktops) > 0 {
		result = append(result, waylandDesktops...)
	}

	// load custom desktops
	customDesktops := listDesktops(pathCustomSessions, Custom)
	if customDesktops != nil && len(customDesktops) > 0 {
		result = append(result, customDesktops...)
	}

	// load custom user desktops
	customUserDesktops := listDesktops(usr.homedir+pathUserCustomSession, UserCustom)
	if customUserDesktops != nil && len(customUserDesktops) > 0 {
		result = append(result, customUserDesktops...)
	}

	return result
}

// List desktops, that could be found on defined path.
func listDesktops(path string, env enEnvironment) []*desktop {
	var result []*desktop

	if fileExists(path) {
		err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
			if !fileInfo.IsDir() && strings.HasSuffix(filePath, ".desktop") {
				d := getDesktop(filePath, env)
				result = append(result, d)
			}
			return nil
		})
		handleErr(err)
	}

	return result
}

// Inits desktop object from .desktop file on defined path.
func getDesktop(path string, env enEnvironment) *desktop {
	d := desktop{env: env, envOrigin: env, isUser: false, path: path}
	if env == Custom {
		d.env = Xorg
	}

	readProperties(path, func(key string, value string) {
		switch key {
		case desktopName:
			d.name = value
		case desktopExec:
			d.exec = value
		case desktopEnvironment:
			d.env = parseEnv(value, constEnvXorg)
		}
	})
	return &d
}

// Parses user-specified configuration from file and returns it as desktop structure.
func loadUserDesktop(homeDir string) (d *desktop, lang string) {
	homeDirConf := homeDir + "/.emptty"
	confDirConf := homeDir + "/.config/emptty"

	for _, confFile := range []string{confDirConf, homeDirConf} {
		if !fileExists(confFile) {
			continue
		}
		d := &desktop{isUser: true, path: confFile, env: Xorg, selection: false}

		err := readProperties(confFile, func(key string, value string) {
			switch key {
			case desktopName:
				d.name = value
			case desktopExec, confCommand:
				d.exec = sanitizeValue(value, "")
			case desktopEnvironment:
				d.env = parseEnv(value, constEnvXorg)
			case desktopLang:
				lang = value
			case confSelection:
				d.selection = parseBool(value, "false")
			}
		})
		handleErr(err)

		if d.exec == "" && !d.selection && !fileIsExecutable(d.path) {
			fmt.Printf("\nMissing Exec value/Using selection and your '%s' is not executable.\n", d.path)
			logPrintf("Missing Exec value/Using selection and your '%s' is not executable.\n", d.path)
			return nil, lang
		}
		if d.selection {
			d.exec = ""
			d.name = ""
		}

		return d, lang
	}

	return nil, lang
}

// Gets index of last used desktop.
func getLastDesktop(usr *sysuser, desktops []*desktop) int {
	l := getUserLastSession(usr)
	if l != nil {
		for i, d := range desktops {
			if d.exec == l.exec && d.env == l.env {
				return i
			}
		}
	}
	return 0
}

// Gets user last session stored in his own home directory.
func getUserLastSession(usr *sysuser) *lastSession {
	path := usr.homedir + pathLastSession
	if fileExists(path) {
		content, err := ioutil.ReadFile(path)
		if err == nil {
			strContent := strings.TrimSpace(string(content))

			arrContent := strings.Split(strContent, ";")
			if len(arrContent) > 0 {
				l := lastSession{}
				l.exec = strings.TrimSpace(arrContent[0])
				if len(arrContent) > 1 {
					l.env = parseEnv(arrContent[1], constEnvXorg)
					return &l
				}
			}
		}
	}
	return nil
}

// Sets Last session for declared sysuser and saves it into user's home directory.
func setUserLastSession(usr *sysuser, d *desktop) {
	doAsUser(usr, func() {
		path := usr.homedir + pathLastSession
		data := fmt.Sprintf("%s;%s\n", d.exec, d.env.stringify())
		err := mkDirsForFile(path, 0744)
		if err != nil {
			logPrint(err)
		}
		err = ioutil.WriteFile(path, []byte(data), 0600)
		if err != nil {
			logPrint(err)
		}
	})
}

// Checks, if user last session file already exists.
func isLastDesktopForSave(usr *sysuser, lastDesktop, currentDesktop *desktop) bool {
	return !fileExists(usr.homedir+pathLastSession) || lastDesktop.exec != currentDesktop.exec || lastDesktop.env != currentDesktop.env
}

// Parse input env and selects corresponding environment.
func parseEnv(env, defaultValue string) enEnvironment {
	switch sanitizeValue(env, defaultValue) {
	case constEnvXorg:
		return Xorg
	case constEnvWayland:
		return Wayland
	}
	return Xorg
}

// Stringify enEnvironment value.
func (e enEnvironment) stringify() string {
	switch e {
	case Xorg:
		return constEnvXorg
	case Wayland:
		return constEnvWayland
	}
	return constEnvXorg
}

// String value of enEnvironment
func (env enEnvironment) string() string {
	strings := []string{constEnvSUndefined, constEnvSXorg, constEnvSWayland, constEnvSCustom, constEnvSUserCustom}
	return strings[env]
}
