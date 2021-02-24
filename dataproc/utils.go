package dataproc

import log "github.com/sirupsen/logrus"

// TODO: Expand logging:
func checkClan(clanTag string) bool {
	log.Info("Entered checkClan()")
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

func checkUint8(intToCheck int64) (uint8, bool) {
	log.Info("Entered checkUint8()")

	if intToCheck < 0 || intToCheck > 255 {
		return uint8(0), false
	} else {
		return uint8(intToCheck), true
	}
}

func checkUint16(intToCheck int64) (uint16, bool) {
	log.Info("Entered checkUint16()")
	if intToCheck < 0 || intToCheck > 65535 {
		return uint16(0), false
	} else {
		return uint16(intToCheck), true
	}
}

func checkUint32(intToCheck int64) (uint32, bool) {
	log.Info("Entered checkUint32()")
	if intToCheck < 0 || intToCheck > 4294967295 {
		return uint32(0), false
	} else {
		return uint32(intToCheck), true
	}
}

func checkUint64(intToCheck int64) (uint64, bool) {
	log.Info("Entered checkUint64()")
	if intToCheck < 0 {
		return uint64(0), false
	} else {
		return uint64(intToCheck), true
	}
}

func checkUint8Float(floatToCheck float64) (uint8, bool) {
	log.Info("Entered checkUint8Float()")

	if floatToCheck < 0 || floatToCheck > 255 {
		return uint8(0), false
	} else {
		return uint8(floatToCheck), true
	}
}

// Helper function checking if a slice contains a string.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
