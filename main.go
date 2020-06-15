package main

const version = "0.2.0"

func main() {
	initLogger()
	handleArgs()
	printMotd()

	conf := loadConfig()
	switchTTY(conf)
	login(conf)
}
