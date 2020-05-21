package main

import (
	"os"
	"path/filepath"
	"strings"
)

// desktop defines structure for display environments and window managers
type desktop struct {
	name string
	exec string
	env  enEnvironment
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

	if _, err := os.Stat(path); err == nil {
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

// Inits desktop object from .desktop fiel on defined path
func getDesktop(path string, env enEnvironment) *desktop {
	d := desktop{env: env}
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
