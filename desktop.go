package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	confEnvironment = "ENVIRONMENT"
	confCommand     = "COMMAND"
	confLang        = "LANG"
)

// desktop defines structure for display environments and window managers
type desktop struct {
	name   string
	exec   string
	env    enEnvironment
	isUser bool
	path   string
}

// Allows to select desktop, which could be selected.
func selectDesktop(uid int) *desktop {
	desktops := listAllDesktops()
	if len(desktops) == 0 {
		log.Fatal("Not found any installed desktop.")
	}

	for true {
		fmt.Printf("\nSelect desktop:\n")
		for i, v := range desktops {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("[%d] %s", i, v.name)
		}
		defaultDesktop := 0
		// TODO: preselected value!
		fmt.Printf(" - [%d]: ", defaultDesktop)

		selection, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		selection = strings.TrimSpace(selection)
		if selection == "" {
			selection = strconv.Itoa(defaultDesktop)
		}

		id, err := strconv.ParseUint(selection, 10, 32)
		if err != nil {
			continue
		}
		if int(id) < len(desktops) {
			// TODO: save last selection!
			return desktops[id]
		}
	}
	return nil
}

// List all installed desktops and return their exec commands.
func listAllDesktops() []*desktop {
	var result []*desktop

	// load Xorg desktops
	xorgDesktops := listDesktops("/usr/share/xsessions/", Xorg)
	if xorgDesktops != nil && len(xorgDesktops) > 0 {
		result = append(result, xorgDesktops...)
	}

	// load Wayland desktops
	waylandDesktops := listDesktops("/usr/share/wayland-sessions/", Xorg)
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

		err := readProperties(confFile, func(key string, value string) error {
			switch key {
			case confCommand:
				d.exec = sanitizeValue(value, "")
				break
			case confEnvironment:
				d.env = parseEnv(value, "xorg")
				break
			case confLang:
				lang = value
			}
			return nil
		})
		handleErr(err)
		return &d, lang
	}

	return nil, lang
}

// Inits desktop object from .desktop fiel on defined path
func getDesktop(path string, env enEnvironment) *desktop {
	d := desktop{env: env, isUser: false, path: path}
	readProperties(path, func(key string, value string) error {
		switch key {
		case "Name":
			d.name = value
			break
		case "Exec":
			d.exec = value
			break
		}
		return nil
	})
	return &d
}
