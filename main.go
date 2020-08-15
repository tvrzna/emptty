package main

const version = "0.2.3"

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
