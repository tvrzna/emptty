package src

import (
	"fmt"
	"os"
)

const version = "0.4.1"

// Shows login screen
func showLoginScreen(conf *config) {
	initLogger(conf)

	printMotd(conf)

	login(conf)
}

// Loads config and shows login screen
func ShowLoginScreen() {
	showLoginScreen(loadConfig())
}

// Handles passed arguments.
func HandleArgs() {
	for _, arg := range os.Args {
		switch arg {
		case "-v", "--version":
			fmt.Printf("emptty %s\nhttps://github.com/tvrzna/emptty\n\nReleased under the MIT License.\n\n", version)
			os.Exit(0)
		case "-d", "--daemon":
			startDaemon()
			os.Exit(0)
		}
	}
}
