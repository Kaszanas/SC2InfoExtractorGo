package dataproc

import log "github.com/sirupsen/logrus"

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
		log.Info("Value does not fit in uint8, returning 0")
		return uint8(0), false
	} else {
		log.Info("Value fits in uint8, converting")
		return uint8(intToCheck), true
	}
}

func checkUint16(intToCheck int64) (uint16, bool) {
	log.Info("Entered checkUint16()")
	if intToCheck < 0 || intToCheck > 65535 {
		log.Info("Value does not fit in uint16, returning 0")
		return uint16(0), false
	} else {
		log.Info("Value fits in uint16, converting")
		return uint16(intToCheck), true
	}
}

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

func checkUint8Float(floatToCheck float64) (uint8, bool) {
	log.Info("Entered checkUint8Float()")

	if floatToCheck < 0 || floatToCheck > 255 {
		log.Info("Value does not fit in uint8, returning 0")
		return uint8(0), false
	} else {
		log.Info("Value fits in uint 8, converting")
		return uint8(floatToCheck), true
	}
}

// Helper function checking if a slice contains a string.
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
