package src

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

const version = "0.11.0"

var buildVersion string

type sessionHandle struct {
	session *commonSession
	auth    authHandle
}

func init() {
	runtime.LockOSThread()
}

// Main handles the functionality of whole application.
func Main() {
	TEST_MODE = false

	processCoreArgs(os.Args)

	conf := loadConfig(loadConfigPath(os.Args))
	processArgs(os.Args, conf)

	fTTY := startDaemon(conf)

	initLogger(conf)
	printMotd(conf)

	login(conf, initSessionHandle())

	stopDaemon(conf, fTTY)
}

// Initialize session handle with common interrupt handler
func initSessionHandle() *sessionHandle {
	h := &sessionHandle{}

	c := make(chan os.Signal, 10)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go handleInterrupt(c, h)

	return h
}

// Catch interrupt signal chan and interrupts Cmd.
func handleInterrupt(c chan os.Signal, h *sessionHandle) {
	<-c
	logPrint("Caught interrupt signal")
	setTerminalEcho(os.Stdout.Fd(), true)
	if h.session != nil && h.session.cmd != nil {
		h.session.interrupted = true
		h.session.cmd.Process.Signal(os.Interrupt)
		h.session.cmd.Wait()
	} else {
		if h.auth != nil {
			h.auth.closeAuth()
		}
		os.Exit(1)
	}
}

// Process core arguments for help and version, because they don't require any further application run
func processCoreArgs(args []string) {
	if contains(args, "-h", "--help") {
		printHelp()
		os.Exit(0)
	}
	if contains(args, "-v", "--version") {
		fmt.Printf("emptty %s\nhttps://github.com/tvrzna/emptty\n\nReleased under the MIT License.\n", getVersion())
		os.Exit(0)
	}
}

// Loads config from path according to arguments
func loadConfigPath(args []string) (configPath string) {
	configPath = pathConfigFile

	for i, arg := range args {
		switch arg {
		case "-c", "--config":
			nextArg(args, i, func(val string) {
				if fileExists(val) {
					configPath = val
				}
			})
			return configPath
		case "-i", "--ignore-config":
			return ""
		}
	}

	return configPath
}

// Process arguments with affection on configuration
func processArgs(args []string, conf *config) {
	for i, arg := range args {
		switch arg {
		case "-t", "--tty":
			nextArg(args, i, func(val string) {
				tty := parseTTY(val, "0")
				if tty > 0 {
					conf.Tty = tty
				} else {
					ttynum := strings.SplitAfterN(val, "tty", 2)
					if len(ttynum) == 2 {
						tty := parseTTY(ttynum[1], "0")
						if tty > 0 {
							conf.Tty = tty
						}
					}
				}
			})
		case "-u", "--default-user":
			nextArg(args, i, func(val string) {
				conf.DefaultUser = val
			})
		case "-d", "--daemon":
			conf.DaemonMode = true
		case "-a", "--autologin":
			conf.Autologin = true
			nextArg(args, i, func(val string) {
				conf.AutologinSession = val
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
	fmt.Printf("  -c, --config PATH\t\tload configuration from specified path\n")
	fmt.Printf("  -i, --ignore-config\t\tskips loading of configuration from file, loads only argument configuration\n")
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
