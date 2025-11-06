package src

import (
	"bufio"
	"fmt"
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
	desktopEnv         = "ENV"
	desktopLang        = "LANG"
	desktopLoginShell  = "LOGINSHELL"
	desktopNames       = "DESKTOPNAMES"
	desktopNoDisplay   = "NODISPLAY"
	desktopHidden      = "HIDDEN"

	constTrue  = "true"
	constFalse = "false"
	constAuto  = "auto"

	pathLastSession       = "/.cache/emptty/last-session"
	pathCustomSessions    = "/etc/emptty/custom-sessions/"
	pathUserCustomSession = "/.config/emptty-custom-sessions/"

	pathLocalWaylandSessions = "/.local/share/wayland-sessions/"
	pathLocalXSessions       = "/.local/share/xsessions/"
)

type enSelection byte

const (

	// Never show selection
	SelectionFalse enSelection = iota

	// Always show selection
	SelectionTrue

	// Show selection only if necessary
	SelectionAuto
)

// desktop defines structure for display environments and window managers.
type desktop struct {
	name         string
	exec         string
	env          enEnvironment
	envOrigin    enEnvironment
	isUser       bool
	path         string
	selection    enSelection
	child        *desktop
	loginShell   string
	desktopNames string
	noDisplay    bool
	hidden       bool
}

// Gets exec path from desktop and returns true, if command allows dbus-launch.
func (d *desktop) getStrExec() (string, bool) {
	if d.selection != SelectionFalse && d.child != nil {
		return d.path + " " + d.child.exec, false
	} else if d.exec != "" {
		return d.exec, true
	}
	return d.path, false
}

// Gets correct desktop name, if is available.
func (d *desktop) getDesktopName() string {
	if d.desktopNames != "" {
		names := strings.Split(d.desktopNames, ":")
		if len(names) > 0 {
			return names[0]
		}
	}
	return d.name
}

// Sets desktop names in expected format
func (d *desktop) setDesktopNames(desktopNames string) {
	val := sanitizeValue(desktopNames, "")
	if val != "" {
		var names []string
		for _, name := range strings.Split(strings.ReplaceAll(val, ";", ":"), ":") {
			if name != "" {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			d.desktopNames = strings.Join(names, ":")
		}
	}
}

// lastSession defines structure for last used session on user login.
type lastSession struct {
	exec string
	env  enEnvironment
}

// Allows to select desktop, which could be selected.
func selectDesktop(usr *sysuser, conf *config, d *desktop) (*desktop, *desktop) {
	allowAutoselectDesktop := d == nil || d.selection == SelectionFalse

	desktops := listAllDesktops(usr, conf.XorgSessionsPath, conf.WaylandSessionsPath)
	if len(desktops) == 0 {
		handleStrErr("Not found any installed desktop.")
	}

	lastDesktop := getLastDesktop(usr, desktops)

	if conf.Autologin && conf.AutologinSession != "" {
		if d := findAutoselectDesktop(conf.AutologinSession, conf.AutologinSessionEnv, desktops); d != nil {
			return d, desktops[lastDesktop]
		}
	}

	if conf.DefaultSession != "" && allowAutoselectDesktop {
		if d := findAutoselectDesktop(conf.DefaultSession, conf.DefaultSessionEnv, desktops); d != nil {
			return d, desktops[lastDesktop]
		}
	}

	// If there is just one desktop and AutoSelection is set or selection is set to Auto, select first desktop
	if len(desktops) == 1 && (conf.AutoSelection || (d != nil && d.selection == SelectionAuto)) {
		return desktops[0], desktops[lastDesktop]
	}

	// Otherwise go through selection process
	for {
		fmt.Printf("\n")
		printDesktops(conf, desktops)
		if conf.VerticalSelection {
			indent := strings.Repeat(" ", conf.IndentSelection)
			fmt.Printf("\n\n%sSelect [%d]: ", indent, lastDesktop)
		} else {
			fmt.Printf("\nSelect [%d]: ", lastDesktop)
		}

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
			return desktops[id], desktops[lastDesktop]
		}
	}
}

// Prints list of desktops on screen
func printDesktops(conf *config, desktops []*desktop) {
	dSeparator := ", "
	eSeparator := " "
	if conf.VerticalSelection {
		indent := strings.Repeat(" ", conf.IndentSelection)
		dSeparator = "\n" + indent
		eSeparator = "\n" + indent
	}

	lastEnv := Undefined
	if conf.VerticalSelection {
		fmt.Print(strings.Repeat(" ", conf.IndentSelection))
	}
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
		if conf.VerticalSelection {
			extraIndent := ""
			if i < 10 {
				extraIndent = " "
			}
			fmt.Printf("%s[%d] %s", extraIndent, i, v.name)
		} else {
			fmt.Printf("[%d] %s", i, v.name)
		}
	}
}

// Finds defined autologinSession in array of desktops by its exec or its name and environment, if defined.
func findAutoselectDesktop(autologinSession string, env enEnvironment, desktops []*desktop) *desktop {
	exec, args := getDesktopBaseExec(autologinSession)
	for _, d := range desktops {
		desktopExec, _ := getDesktopBaseExec(d.exec)
		if (exec == desktopExec || strings.EqualFold(autologinSession, d.name)) &&
			(env == Undefined || env == d.env) {
			if args != "" {
				d.exec = d.exec + " " + args
			}
			return d
		}
	}
	return nil
}

// Gets base executable name of desktop
func getDesktopBaseExec(exec string) (string, string) {
	parts := strings.Split(strings.TrimSpace(exec), "/")
	value := strings.TrimSpace(parts[len(parts)-1])
	sep := strings.Index(value, " ")
	if sep > -1 {
		return value[:sep], value[sep+1:]
	}
	return value, ""
}

// List all installed desktops and return their exec commands.
func listAllDesktops(usr *sysuser, pathXorgDesktops, pathWaylandDesktops string) []*desktop {
	var result []*desktop

	// load Xorg desktops
	result = append(result, listDesktops(Xorg, pathXorgDesktops, usr.homedir+pathLocalXSessions)...)

	// load Wayland desktops
	result = append(result, listDesktops(Wayland, pathWaylandDesktops, usr.homedir+pathLocalWaylandSessions)...)

	// load custom desktops
	result = append(result, listDesktops(Custom, pathCustomSessions)...)

	// load custom user desktops
	result = append(result, listDesktops(UserCustom, usr.homedir+pathUserCustomSession)...)

	return result
}

// List desktops, that could be found on defined paths.
func listDesktops(env enEnvironment, paths ...string) []*desktop {
	var result []*desktop

	for _, path := range paths {
		if strings.HasSuffix(path, "/") {
			path += "/"
		}

		if fileExists(path) {
			err := filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
				if !fileInfo.IsDir() && strings.HasSuffix(filePath, ".desktop") {
					d := getDesktop(filePath, env)
					if !d.noDisplay && !d.hidden {
						result = append(result, d)
					}
				}
				return nil
			})
			handleErr(err)
		}

	}
	return result
}

// Inits desktop object from .desktop file on defined path.
func getDesktop(path string, env enEnvironment) *desktop {
	d := desktop{env: env, envOrigin: env, isUser: false, path: path}
	if env == Custom || env == UserCustom {
		d.env = defaultEnvValue
	}

	readProperties(path, func(key string, value string) {
		switch key {
		case desktopName:
			d.name = value
		case desktopExec:
			d.exec = value
		case desktopEnvironment, desktopEnv:
			d.env = parseEnv(value, defaultEnv())
		case desktopNames:
			d.setDesktopNames(value)
		case desktopNoDisplay:
			d.noDisplay = parseBool(value, "false")
		case desktopHidden:
			d.hidden = parseBool(value, "false")
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
		d := &desktop{isUser: true, path: confFile, env: defaultEnvValue, selection: SelectionFalse}

		err := readPropertiesWithSupport(confFile, func(key string, value string) {
			switch key {
			case desktopName:
				d.name = value
			case desktopExec, confCommand:
				d.exec = sanitizeValue(value, "")
			case desktopEnvironment, desktopEnv:
				d.env = parseEnv(value, defaultEnv())
			case desktopLang:
				lang = value
			case confSelection:
				d.selection = parseSelection(value, "false")
			case desktopLoginShell:
				d.loginShell = sanitizeValue(value, "")
			case desktopNames:
				d.setDesktopNames(value)
			}
		}, true)
		handleErr(err)

		if d.selection != SelectionFalse {
			d.exec = ""
			d.name = ""
			d.desktopNames = ""
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
		if content, err := os.ReadFile(path); err == nil {
			arrContent := strings.Split(strings.TrimSpace(string(content)), ";")
			l := lastSession{}
			l.exec = strings.TrimSpace(arrContent[0])
			if len(arrContent) > 1 {
				l.env = parseEnv(arrContent[1], defaultEnv())
				return &l
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
		if err := mkDirsForFile(path, 0744); err != nil {
			logPrint(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			logPrint(err)
		}
	})
}

// Checks, if user last session file already exists.
func isLastDesktopForSave(usr *sysuser, lastDesktop, currentDesktop *desktop) bool {
	return !fileExists(usr.homedir+pathLastSession) || lastDesktop.exec != currentDesktop.exec || lastDesktop.env != currentDesktop.env
}

// Parse input selection
func parseSelection(selection, defaultValue string) enSelection {
	switch strings.ToLower(sanitizeValue(selection, defaultValue)) {
	case constTrue:
		return SelectionTrue
	case constFalse:
		return SelectionFalse
	case constAuto:
		return SelectionAuto
	}
	return SelectionFalse
}
