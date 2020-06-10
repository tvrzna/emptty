package main

const version = "0.2.0"

func main() {
	handleArgs()

	printMotd()

	conf := loadConfig()

	switchTTY(conf.tty)

	login(conf)
}
