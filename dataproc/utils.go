package dataproc

import (
	log "github.com/sirupsen/logrus"
)

// contains is a helper function checking if a slice contains a string.
func contains(s []string, str string) bool {
	log.Info("Entered contains()")

	for _, v := range s {
		if v == str {
			log.Debug("Slice contains supplied string, returning true")
			return true
		}
	}

	log.Info("Slice does not contain supplied string, returning false")
	return false
}
