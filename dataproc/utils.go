package dataproc

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// checkClan verifies if a player is in a clan.
func checkClan(clanTag string) bool {
	log.Info("Entered checkClan()")
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

// checkUint8 verifies if an int64 can fit into uint8
// and returns a converted variable and a boolean.
func checkUint8(intToCheck int64) (uint8, bool) {
	log.Info("Entered checkUint8()")

	if intToCheck < 0 || intToCheck > 255 {
		log.Info("Value does not fit in uint8, returning 0")
		return uint8(0), false
	} else {
		log.Info("Value fits in uint8, converting")
		return uint8(intToCheck), true
	}
}

// checkUint32 verifies if an int64 can fit into uint64
// and returns a converted variable and a boolean.
func checkUint32(intToCheck int64) (uint32, bool) {
	log.Info("Entered checkUint32()")
	if intToCheck < 0 || intToCheck > 4294967295 {
		log.Info("Value does not fit in uint32, returning 0")
		return uint32(0), false
	} else {
		log.Info("Value fits in uint32, converting")
		return uint32(intToCheck), true
	}
}

// checkUint64 verifies if an int64 can fit into uint32
// and returns a converted variable and a boolean.
func checkUint64(intToCheck int64) (uint64, bool) {
	log.Info("Entered checkUint64()")

	// Hard to verify as int64 will always fit in uint64!!!
	if intToCheck < 0 || intToCheck > 9223372036854775807 {
		log.Info("Value does not fit in uint64, returning 0")
		return uint64(0), false
	} else {
		log.Info("Value fits in uint64, converting")
		return uint64(intToCheck), true
	}
}

// contains is a helper function checking if a slice contains a string.
func contains(s []string, str string) bool {
	log.Info("Entered contains()")

	for _, v := range s {
		if v == str {
			log.Info("Slice contains supplied string, returning true")
			return true
		}
	}

	log.Info("Slice does not contain supplied string, returning false")
	return false
}

func replaceNewlinesSplitData(input string) []string {
	replacedNewlines := strings.ReplaceAll(input, "\r\n", "\n")
	splitFile := strings.Split(replacedNewlines, "\n")

	return splitFile
}
