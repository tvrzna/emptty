package src

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	confEnvironment = "ENVIRONMENT"
	confCommand     = "COMMAND"

	desktopExec        = "EXEC"
	desktopName        = "NAME"
	desktopEnvironment = "ENVIRONMENT"

	constEnvXorg    = "xorg"
	constEnvWayland = "wayland"

	pathLastSession       = "/.cache/emptty/last-session"
	pathXorgSessions      = "/usr/share/xsessions/"
	pathWaylandSessions   = "/usr/share/wayland-sessions/"
	pathCustomSessions    = "/etc/emptty/custom-sessions/"
	pathUserCustomSession = "/.config/emptty-custom-sessions/"
)

// enEnvironment defines possible Environments.
type enEnvironment int

const (
	// Xorg represents Xorg environment
	Xorg enEnvironment = iota + 1

	// Wayland represents Wayland environment
	Wayland

	Custom
)

// desktop defines structure for display environments and window managers.
type desktop struct {
	name   string
	exec   string
	env    enEnvironment
	isUser bool
	path   string
}

// lastSession defines structure for last used session on user login.
type lastSession struct {
	uid  int
	exec string
	env  enEnvironment
}

// Allows to select desktop, which could be selected.
func selectDesktop(usr *sysuser, conf *config) *desktop {
	desktops := listAllDesktops(usr)
	if len(desktops) == 0 {
		handleStrErr("Not found any installed desktop.")
	}

	lastDesktop := getLastDesktop(usr, desktops)

	if conf.autologin && conf.autologinSession != "" {
		for _, d := range desktops {
			if strings.HasSuffix(d.exec, conf.autologinSession) {
				if isLastDesktopForSave(usr, desktops[lastDesktop], d) {
					setUserLastSession(usr, d)
				}
				return d
			}
		}
	}

	for true {
		fmt.Printf("\n")
		for i, v := range desktops {
			if i > 0 {
				if conf.verticalSelection {
					fmt.Print("\n")
				} else {
					fmt.Print(", ")
				}
			}
			fmt.Printf("[%d] %s", i, v.name)
		}
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

// List all installed desktops and return their exec commands.
func listAllDesktops(usr *sysuser) []*desktop {
	var result []*desktop

	// load Xorg desktops
	xorgDesktops := listDesktops(pathXorgSessions, Xorg)
	if xorgDesktops != nil && len(xorgDesktops) > 0 {
		result = append(result, xorgDesktops...)
	}

	// load Wayland desktops
	waylandDesktops := listDesktops(pathWaylandSessions, Wayland)
	if waylandDesktops != nil && len(waylandDesktops) > 0 {
		result = append(result, waylandDesktops...)
	}

	// load custom desktops
	customDesktops := listDesktops(pathCustomSessions, Custom)
	if customDesktops != nil && len(customDesktops) > 0 {
		result = append(result, customDesktops...)
	}

	// load custom user desktops
	customUserDesktops := listDesktops(usr.homedir+pathUserCustomSession, Custom)
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
	d := desktop{env: env, isUser: false, path: path}
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
func loadUserDesktop(homeDir string) (*desktop, string) {
	homeDirConf := homeDir + "/.emptty"
	confDirConf := homeDir + "/.config/emptty"

	var lang string
	for _, confFile := range []string{confDirConf, homeDirConf} {
		if fileExists(confFile) {
			d := &desktop{isUser: true, path: confFile, env: Xorg}

			err := readProperties(confFile, func(key string, value string) {
				switch key {
				case desktopName:
					d.name = value
				case desktopExec, confCommand:
					d.exec = sanitizeValue(value, "")
				case desktopEnvironment:
					d.env = parseEnv(value, constEnvXorg)
				case confLang:
					lang = value
				}
			})
			handleErr(err)

			if d.exec == "" && !fileIsExecutable(d.path) {
				fmt.Printf("\nMissing Exec value and your '%s' is not executable.\n", d.path)
				log.Printf("Missing Exec value and your '%s' is not executable.\n", d.path)
				return nil, lang
			}

			return d, lang
		}
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
	currentUser, _ := user.Current()
	previousUser := getSysuser(currentUser)

	setFsUser(usr)

	path := usr.homedir + pathLastSession
	data := fmt.Sprintf("%s;%s\n", d.exec, stringifyEnv(d.env))
	err := mkDirsForFile(path, 0744)
	if err != nil {
		log.Print(err)
	}
	err = ioutil.WriteFile(path, []byte(data), 0600)
	if err != nil {
		log.Print(err)
	}

	setFsUser(previousUser)
}

// Checks, if user last session file already exists.
func isLastDesktopForSave(usr *sysuser, lastDesktop *desktop, currentDesktop *desktop) bool {
	return !fileExists(usr.homedir+pathLastSession) || lastDesktop.exec != currentDesktop.exec || lastDesktop.env != currentDesktop.env
}

// Parse input env and selects corresponding environment.
func parseEnv(env string, defaultValue string) enEnvironment {
	switch sanitizeValue(env, defaultValue) {
	case constEnvXorg:
		return Xorg
	case constEnvWayland:
		return Wayland
	}
	return Xorg
}

// Stringify enEnvironment value.
func stringifyEnv(env enEnvironment) string {
	switch env {
	case Xorg:
		return constEnvXorg
	case Wayland:
		return constEnvWayland
	}
	return constEnvXorg
}
