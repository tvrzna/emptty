package src

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

// TEST_MODE Defines if logging is in test mode
var TEST_MODE bool

const (
	pathLogFileNull      = "/dev/null"
	pathLogFileOldSuffix = ".old"

	constTTYplaceholder = "[TTY_NUMBER]"

	constLogDefault   = "default"
	constLogRotate    = "rotate"
	constLogAppending = "appending"
	constLogDisabled  = "disabled"
)

// enLogging defines possible option how to handle configuration.
type enLogging byte

const (
	// Rotate represents saving into new file and backing up older with suffix
	Rotate enLogging = iota + 1

	// Appending represents saving all logs into same file
	Appending

	// Disabled represents disabled logging
	Disabled
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
		waitForReturnToExit(1)
	}
}

// Initialize logger to file defined by pathLogFile.
func initLogger(conf *config) {
	f, err := prepareLogFile(conf.LoggingFile, conf.strTTY(), conf.Logging)
	if err == nil {
		log.SetOutput(f)
	}
}

// Initialize logger to file for session-errors.
func initSessionErrorLogger(conf *config) (*os.File, error) {
	return prepareLogFile(conf.SessionErrLogFile, conf.strTTY(), conf.SessionErrLog)
}

// Prepares logging file according to defined configuration.
func prepareLogFile(path, tty string, method enLogging) (*os.File, error) {
	logFilePath := strings.ReplaceAll(path, constTTYplaceholder, tty)

	if method == Rotate && logFilePath != pathLogFileNull {
		// Temporary workaround to allow create new folder
		backupFileIfNotFolder(logFilePath)

		if err := mkDirsForFile(logFilePath, 0755); err != nil {
			return nil, err
		}
		if fileExists(logFilePath) {
			os.Remove(logFilePath + pathLogFileOldSuffix)
			os.Rename(logFilePath, logFilePath+pathLogFileOldSuffix)
		}
	} else if method == Disabled {
		logFilePath = pathLogFileNull
	}

	return os.OpenFile(logFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
}

// Temporal solution to avoid issues with names of logging folder, if there is already file with same name.
func backupFileIfNotFolder(path string) {
	fileName := path[:strings.LastIndex(path, "/")]
	if f, err := os.Stat(fileName); err == nil && f != nil && !f.IsDir() {
		os.Remove(fileName + pathLogFileOldSuffix)
		os.Rename(fileName, fileName+pathLogFileOldSuffix)
	}
}

// Parse logging option
func parseLogging(strLogging, defaultValue string) enLogging {
	switch strings.ToLower(sanitizeValue(strLogging, defaultValue)) {
	case constLogDisabled:
		return Disabled
	case constLogAppending:
		return Appending
	case constLogDefault, constLogRotate:
		return Rotate
	}
	return Rotate
}

// Stringify logging option
func (l *enLogging) stringify() string {
	strings := []string{"", constLogDisabled, constLogAppending, constLogRotate}
	return strings[int(*l)]
}
