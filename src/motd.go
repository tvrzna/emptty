package src

import (
	"bufio"
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
		cmd := exec.Command(conf.DynamicMotdPath)
		stdout, _ := cmd.StdoutPipe()
		cmd.Start()

		buf := bufio.NewReader(stdout)
		for {
			line, _, err := buf.ReadLine()
			if err == nil {
				fmt.Println(string(line))
			} else {
				break
			}
		}
		cmd.Wait()

		revertColors(conf)
		return true
	}
	return false
}

// Prints static motd. If something was printed, returns true.
func printStaticMotd(conf *config) bool {
	if fileExists(conf.MotdPath) {
		motd, err := os.ReadFile(conf.MotdPath)
		if err != nil {
			logPrint(err)
			return false
		}
		if len(motd) > 0 {
			fmt.Println(revertColorEscaping(string(motd)))
			revertColors(conf)
		}
		return true
	}
	return false
}

// Reverts output color to default.
func revertColors(conf *config) {
	if conf.DaemonMode {
		setColors(conf.FgColor, conf.BgColor)
	} else {
		resetColors()
	}
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
