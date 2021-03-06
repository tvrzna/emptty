package src

import (
	"fmt"
	"os"
	"strings"
)

const version = "0.5.2"

var buildVersion string

// Main handles the functionality of whole application.
func Main() {
	TEST_MODE = false

	if contains(os.Args, "-h") || contains(os.Args, "--help") {
		printHelp()
		os.Exit(0)
	}
	if contains(os.Args, "-v") || contains(os.Args, "--version") {
		fmt.Printf("emptty %s\nhttps://github.com/tvrzna/emptty\n\nReleased under the MIT License.\n", getVersion())
		os.Exit(0)
	}

	conf := loadConfig(pathConfigFile)
	processArgs(os.Args, conf)

	var fTTY *os.File
	if conf.daemonMode {
		fTTY = startDaemon(conf)
	}

	initLogger(conf)
	printMotd(conf)
	login(conf)

	if conf.daemonMode {
		stopDaemon(conf, fTTY)
	}
}

// Process arguments with affection on configuration
func processArgs(args []string, conf *config) {
	for i, arg := range args {
		switch arg {
		case "-t", "--tty":
			nextArg(args, i, func(val string) {
				tty := parseTTY(val, "0")
				if tty > 0 {
					conf.tty = tty
				}
			})
		case "-u", "--default-user":
			nextArg(args, i, func(val string) {
				conf.defaultUser = val
			})
		case "-d", "--daemon":
			conf.daemonMode = true
		case "-a", "--autologin":
			conf.autologin = true
			nextArg(args, i, func(val string) {
				conf.autologinSession = val
			})
		}
	}
}

// Gets next argument, if available
func nextArg(args []string, i int, callback func(value string)) {
	if callback != nil && len(args) > i+1 {
		val := sanitizeValue(args[i+1], "")
		if !strings.HasPrefix(val, "-") {
			callback(args[i+1])
		}
	}
}

// Prints help
func printHelp() {
	fmt.Println("Usage: emptty [options]")
	fmt.Println("Options:")
	fmt.Printf("  -h, --help\t\t\tprint this help\n")
	fmt.Printf("  -v, --version\t\t\tprint version\n")
	fmt.Printf("  -d, --daemon\t\t\tstart in daemon mode\n")
	fmt.Printf("  -t, --tty NUMBER\t\toverrides configured TTY number\n")
	fmt.Printf("  -u, --default-user USER_NAME\toverrides configured Default User\n")
	fmt.Printf("  -a, --autologin [SESSION]\toverrides configured autologin to true and if next argument is defined, it defines also Autologin Session.\n")
}

// Gets current version
func getVersion() string {
	tags := strings.Builder{}
	for _, tag := range []string{tagPam, tagUtmp, tagXlib} {
		if tags.Len() > 0 {
			tags.WriteString(", ")
		}
		tags.WriteString(tag)
	}
	if buildVersion != "" {
		if tags.Len() == 0 {
			return buildVersion[1:]
		}
		return buildVersion[1:] + " (" + tags.String() + ")"
	}
	return version
}
