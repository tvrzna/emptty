package main

const version = "0.3.0"

func main() {
	handleArgs()

	showLoginScreen(loadConfig())
}

// Shows login screen
func showLoginScreen(conf *config) {
	initLogger(conf)

	printMotd(conf)

	login(conf)
}
