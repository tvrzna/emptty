package src

import (
	"path/filepath"
	"time"
)

const (
	cardDevicePathPrefix = "/dev/dri"
	maxWaitForCardDevice = 1 * time.Minute
	waitBetweenChecks    = 1 * time.Second
)

func waitForCardDevice(conf *config) {
	if conf.WaitCardDeviceCount < 1 {
		return
	}

	logPrintf("Waiting %s for required card device(s)", maxWaitForCardDevice)
	deadline := time.Now().Add(maxWaitForCardDevice)
	for time.Now().Before(deadline) {
		cardList, err := filepath.Glob(cardDevicePathPrefix + "/card*")
		if err != nil {
			logPrintf("Failed to search for card devices: %v", err)
			return
		}

		if conf.WaitCardDeviceCount <= len(cardList) {
			logPrintf("Required card device(s) found")
			return
		}

		time.Sleep(waitBetweenChecks)
	}

	logPrintf("Required card device(s) not found, session may fail to launch")
}
