package main

const version = "0.1.0"

func main() {
	handleArgs()

	printMotd()

	conf := loadConfig()

	switchTTY(conf.tty)

	login(conf)
}
