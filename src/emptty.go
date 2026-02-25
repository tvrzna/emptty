package src

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

const (
	version = "0.15.0"

	builtinCmdHelpString = `
Available commands:
  :help, :?			print this help
  :poweroff, :shutdown		process poweroff command
  :reboot			process reboot command
  :suspend, :zzz		process suspend command
`
	terminalHelpString = `Usage: emptty [options]
Options:
  -h, --help			print this help
  -v, --version			print version
  -d, --daemon			start in daemon mode
  -c, --config PATH		load configuration from specified path
  -C, --print-config	prints currently loaded configuration
  -i, --ignore-config		skips loading of configuration from file, loads only argument configuration
  -t, --tty NUMBER		overrides configured TTY number
  -u, --default-user USER_NAME	overrides configured Default User
  -a, --autologin [SESSION]	overrides configured autologin to true and if next argument is defined, it defines also Autologin Session
`
)

var buildVersion string
var errPrintCommandHelp = errors.New("just print help")

type sessionHandle struct {
	session     *commonSession
	auth        authHandle
	interrupted bool
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

	if command := login(conf, initSessionHandle()); command != "" {
		processCommand(command, conf, nil, false)
	}

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
	h.interrupted = true

	if h.session != nil && h.session.cmd != nil {
		h.session.interrupted = true
		if err := h.session.cmd.Process.Signal(os.Interrupt); err != nil {
			logPrint("Application not responding to signal")
		} else {
			// Only attempt clean shutdown if cmd is still
			// responsive
			h.session.cmd.Wait()
		}
	}
	// Always attempt to close active auth instances before exit
	if h.auth != nil {
		h.auth.closeAuth()
	}
	os.Exit(1)
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
	printConfig := false

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
		case "-C", "--print-config":
			printConfig = true
		}
	}

	if printConfig {
		conf.printConfig()
		os.Exit(0)
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
	fmt.Print(terminalHelpString)
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

func formatCommand(command string) string {
	return strings.ReplaceAll(command, "\x1b", "")[1:]
}

func shouldProcessCommand(input string, conf *config) bool {
	return conf.AllowCommands && strings.HasPrefix(strings.ReplaceAll(input, "\x1b", ""), ":")
}

// Process commands input in login buffer
func processCommand(command string, c *config, auth authHandle, continuable bool) error {
	switch command {
	case "help", "?":
		fmt.Print(builtinCmdHelpString)
		if continuable {
			return errPrintCommandHelp
		}
		waitForReturnToExit(0)
	case "poweroff", "shutdown":
		if continuable && auth != nil {
			auth.closeAuth()
		}
		if err := processCommandAsCmd(c.CmdPoweroff); err == nil {
			waitForReturnToExit(0)
		} else {
			handleErr(err)
		}
	case "reboot":
		if continuable && auth != nil {
			auth.closeAuth()
		}
		if err := processCommandAsCmd(c.CmdReboot); err == nil {
			waitForReturnToExit(0)
		} else {
			handleErr(err)
		}
	case "suspend", "zzz":
		if continuable && auth != nil {
			auth.closeAuth()
		}
		variants := []string{
			"zzz",
			"systemctl suspend",
			"loginctl suspend",
		}
		if c.CmdSuspend != "" {
			variants = append([]string{c.CmdSuspend}, variants...)
		}

		var err error
		for _, v := range variants {
			if err = processCommandAsCmd(v); err != nil {
				continue
			} else {
				break
			}
		}

		if err == nil {
			waitForReturnToExit(0)
		} else {
			handleErr(err)
		}
	default:
		err := fmt.Errorf("Unknown command '%s'", command)
		if continuable {
			return err
		}
		handleErr(err)
	}
	return nil
}
