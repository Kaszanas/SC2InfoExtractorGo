package cleanup

import log "github.com/sirupsen/logrus"

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
