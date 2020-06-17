package main

const version = "0.2.0"

func main() {
	handleArgs()

	showLoginScreen(loadConfig())
}

// Shows login screen
func showLoginScreen(conf *config) {
	initLogger()

	printMotd()

	login(conf)
}
