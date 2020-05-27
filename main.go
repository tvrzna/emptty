package main

import (
	"fmt"
)

const version = "0.1.0"
const motd = `┌─┐┌┬┐┌─┐┌┬┐┌┬┐┬ ┬
├┤ │││├─┘ │  │ └┬┘
└─┘┴ ┴┴   ┴  ┴  ┴   ` + version

var conf *config

func main() {
	handleArgs()

	fmt.Printf("%s\n\n", motd)

	conf = loadConfig()

	switchTTY(conf.tty)

	login()
}
