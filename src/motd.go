package src

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultMotd = `┌─┐┌┬┐┌─┐┌┬┐┌┬┐┬ ┬
├┤ │││├─┘ │  │ └┬┘
└─┘┴ ┴┴   ┴  ┴  ┴   ` + version
)

// Prints dynamic motd, if configured; otherwise prints motd, if pathMotd exists; otherwise it prints default motd.
func printMotd(conf *config) {
	if !conf.PrintMotd {
		return
	}
	if !printDynamicMotd(conf) {
		if !printStaticMotd(conf) {
			printDefaultMotd()
		}
	}
}

// Prints dynamic motd. If something was printed, returns true.
func printDynamicMotd(conf *config) bool {
	if conf.DynamicMotd && fileIsExecutable(conf.DynamicMotdPath) {
		motd, err := exec.Command(conf.DynamicMotdPath).Output()
		return printCommonMotd(conf, motd, err)
	}
	return false
}

// Prints static motd. If something was printed, returns true.
func printStaticMotd(conf *config) bool {
	if fileExists(conf.MotdPath) {
		motd, err := os.ReadFile(conf.MotdPath)
		return printCommonMotd(conf, motd, err)
	}
	return false
}

// Handles common part of printing motd
func printCommonMotd(conf *config, motd []byte, err error) bool {
	if err != nil {
		logPrint(err)
		return false
	}
	if len(motd) > 0 {
		fmt.Println(revertColorEscaping(string(motd)))
		if conf.DaemonMode {
			setColors(conf.FgColor, conf.BgColor)
		} else {
			resetColors()
		}
	}
	return true
}

// Prints default motd.
func printDefaultMotd() {
	fmt.Printf("%s\n\n", defaultMotd)
}

// Reverts escaped color definitions to real color values.
func revertColorEscaping(value string) string {
	if value != "" {
		result := strings.ReplaceAll(value, "\\x1b", "\x1b")
		result = strings.ReplaceAll(result, "\\033", "\x1b")
		return result
	}
	return value
}

// Sets defined colors.
func setColors(fg, bg string) {
	color := ""

	if fg != "" {
		color += fg
	}
	if fg != "" && bg != "" {
		color += ";"
	}
	if bg != "" {
		color += bg
	}

	if fg == "" && bg == "" {
		color = "0"
	}
	fmt.Print("\x1b[0;" + color + "m\n")
}

// Resets colors to default.
func resetColors() {
	setColors("", "")
}
