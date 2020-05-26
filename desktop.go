package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	confEnvironment = "ENVIRONMENT"
	confCommand     = "COMMAND"

	constEnvXorg    = "xorg"
	constEnvWayland = "wayland"

	pathLastSessions    = "/etc/emptty/last-sessions"
	pathXorgSessions    = "/usr/share/xsessions/"
	pathWaylandSessions = "/usr/share/wayland-sessions/"
)

// desktop defines structure for display environments and window managers
type desktop struct {
	name   string
	exec   string
	env    enEnvironment
	isUser bool
	path   string
}

// lastSession defines structure for last used session on user login
type lastSession struct {
	uid  int
	exec string
	env  enEnvironment
}

// Allows to select desktop, which could be selected.
func selectDesktop(uid int) *desktop {
	desktops := listAllDesktops()
	if len(desktops) == 0 {
		log.Fatal("Not found any installed desktop.")
	}

	lastSessions := loadLastSessions()

	for true {
		fmt.Printf("\n")
		for i, v := range desktops {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("[%d] %s", i, v.name)
		}
		lastDesktop := getLastDesktop(uid, desktops, lastSessions)
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
			setLastSession(uid, d, lastSessions)
			return d
		}
	}
	return nil
}

// List all installed desktops and return their exec commands.
func listAllDesktops() []*desktop {
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

	return result
}

// List desktops, that could be found on defined path
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

// Parses user-specified configuration from file and returns it as desktop structure
func loadUserDesktop(homeDir string) (*desktop, string) {
	confFile := homeDir + "/.emptty"

	var lang string
	if fileExists(confFile) {
		d := desktop{isUser: true, path: confFile, env: Xorg}

		err := readProperties(confFile, func(key string, value string) {
			switch key {
			case confCommand:
				d.exec = sanitizeValue(value, "")
			case confEnvironment:
				d.env = parseEnv(value, constEnvXorg)
			case confLang:
				lang = value
			}
		})
		handleErr(err)
		return &d, lang
	}

	return nil, lang
}

// Inits desktop object from .desktop fiel on defined path
func getDesktop(path string, env enEnvironment) *desktop {
	d := desktop{env: env, isUser: false, path: path}
	readProperties(path, func(key string, value string) {
		switch key {
		case "Name":
			d.name = value
		case "Exec":
			d.exec = value
		}
	})
	return &d
}

// Gets index of last used desktop.
func getLastDesktop(uid int, desktops []*desktop, lastSessions []*lastSession) int {
	l := getLastSession(uid, lastSessions)
	if l != nil {
		for i, d := range desktops {
			if d.exec == l.exec && d.env == l.env {
				return i
			}
		}
	}
	return 0
}

// Gets Last Session of declared uid.
func getLastSession(uid int, lastSessions []*lastSession) *lastSession {
	if lastSessions != nil {
		for _, session := range lastSessions {
			if session.uid == uid {
				return session
			}
		}
	}
	return nil
}

// Sets Last session for declared uid and saves all to file.
func setLastSession(uid int, d *desktop, lastSessions []*lastSession) {
	session := getLastSession(uid, lastSessions)

	if session == nil {
		lastSessions = append(lastSessions, &lastSession{uid: uid, exec: d.exec, env: d.env})
	} else {
		session.exec = d.exec
		session.env = d.env
	}

	saveLastSessions(lastSessions)
}

// Load all last sessions from file.
func loadLastSessions() []*lastSession {
	var result []*lastSession
	if fileExists(pathLastSessions) {
		readProperties(pathLastSessions, func(key string, value string) {
			l := lastSession{}

			uid, err := strconv.ParseInt(key, 10, 32)
			if err != nil {
				return
			}
			l.uid = int(uid)

			arrValue := strings.Split(value, ";")
			l.exec = arrValue[0]
			l.env = parseEnv(arrValue[1], constEnvXorg)

			result = append(result, &l)
		})
	}
	return result
}

// Save all last sessions to file.
func saveLastSessions(lastSessions []*lastSession) {
	var arr []string
	for _, s := range lastSessions {
		arr = append(arr, fmt.Sprintf("%d=%s;%s", s.uid, s.exec, stringifyEnv(s.env)))
	}
	resultStr := strings.Join(arr, "\n")

	err := ioutil.WriteFile(pathLastSessions, []byte(resultStr), 0600)
	if err != nil {
		log.Print(err)
	}
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

// Stringify enEnvironment value
func stringifyEnv(env enEnvironment) string {
	switch env {
	case Xorg:
		return constEnvXorg
	case Wayland:
		return constEnvWayland
	}
	return constEnvXorg
}
