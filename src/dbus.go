package src

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const (
	dbusSessionBusAddress = "DBUS_SESSION_BUS_ADDRESS"
	dbusSessionBusPid     = "DBUS_SESSION_BUS_PID"
)

type dbus struct {
	pid     int
	address string
}

// Launches dbus-launch to start daemon, parses its address and pid for further usage in environmentals.
func (d *dbus) launch(usr *sysuser) {
	logPrint("Starting dbus-launch")
	dbusOutput := runSimpleCmdAsUser(usr, "dbus-launch")
	if dbusOutput == "" {
		logPrint("No output from dbus-launch")
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(dbusOutput))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		readPropertyLine(line, func(key, value string) {
			switch key {
			case dbusSessionBusPid:
				var err error
				d.pid, err = strconv.Atoi(value)
				if err != nil {
					logPrint(err)
				}
			case dbusSessionBusAddress:
				d.address = value
				usr.setenv(dbusSessionBusAddress, value)
			}
		}, false)
	}
	if scanner.Err() != nil {
		logPrint("Reading output from dbus-launch error: ", scanner.Err())
	}
}

// Interrupts dbus by defined pid.
func (d *dbus) interrupt() {
	if d.pid <= 0 {
		logPrint("Trying to interrupt non-existing process")
		return
	}
	proc, err := os.FindProcess(d.pid)
	if err != nil {
		logPrint(err)
	}
	if proc != nil {
		logPrint("Interrupting dbus-daemon, pid: ", d.pid)
		if err := proc.Signal(os.Interrupt); err != nil {
			logPrint("Could not interrupt dbus-daemon (pid: ", d.pid, ")")
		}
	}
}
