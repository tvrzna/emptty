package src

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
)

// TEST_MODE Defines if logging is in test mode
var TEST_MODE bool

const (
	pathLogFileNull      = "/dev/null"
	pathLogFile          = "/var/log/emptty"
	pathLogSessErrFile   = "/var/log/emptty-session-errors"
	pathLogFileOldSuffix = ".old"

	constLogDefault   = "default"
	constLogAppending = "appending"
	constLogDisabled  = "disabled"
)

// Log simple information
func logPrint(v ...interface{}) {
	log.Print(v...)
}

// Log simple information with format
func logPrintf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

// Log fatal information
func logFatal(v ...interface{}) {
	log.Fatal(v...)
}

// Handles error passed as string and calls handleErr function.
func handleStrErr(err string) {
	if err != "" {
		handleErr(errors.New(err))
	}
}

// If error is not nil, otherwise it prints error, waits for user input and then exits the program.
func handleErr(err error) {
	if err != nil {
		logPrint(err)
		fmt.Printf("Error: %s\n", err)
		fmt.Printf("\nPress Enter to continue...")
		if !TEST_MODE {
			bufio.NewReader(os.Stdin).ReadString('\n')
			os.Exit(1)
		}
	}
}

// Initialize logger to file defined by pathLogFile.
func initLogger(conf *config) {
	f, err := prepareLogFile(conf.loggingFile, pathLogFile, conf.logging)
	if err == nil {
		log.SetOutput(f)
	}
}

// Initialize logger to file for session-errors.
func initSessionErrorLogger(conf *config) (*os.File, error) {
	return prepareLogFile(conf.sessionErrLogFile, pathLogSessErrFile, conf.sessionErrLog)
}

// Prepares logging file according to defined configuration.
func prepareLogFile(path string, defaultPath string, method enLogging) (*os.File, error) {
	logFilePath := defaultPath
	if path != "" {
		logFilePath = path
	}

	if method == Default && logFilePath != pathLogFileNull {
		if fileExists(logFilePath) {
			os.Remove(logFilePath + pathLogFileOldSuffix)
			os.Rename(logFilePath, logFilePath+pathLogFileOldSuffix)
		}
	} else if method == Disabled {
		logFilePath = pathLogFileNull
	}

	return os.OpenFile(logFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
}

// Parse logging option
func parseLogging(strLogging string, defaultValue string) enLogging {
	val := sanitizeValue(strLogging, defaultValue)
	switch val {
	case constLogDisabled:
		return Disabled
	case constLogAppending:
		return Appending
	case constLogDefault:
		return Default
	}
	return Default
}
