package dataproc

// TODO: Introduce logging.
func checkClan(clanTag string) bool {
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

func checkUint8(intToCheck int64) (uint8, bool) {

	if intToCheck < 0 || intToCheck > 255 {
		return uint8(0), false
	} else {
		return uint8(intToCheck), true
	}
}

func checkUint16(intToCheck int64) (uint16, bool) {
	if intToCheck < 0 || intToCheck > 65535 {
		return uint16(0), false
	} else {
		return uint16(intToCheck), true
	}
}

func checkUint32(intToCheck int64) (uint32, bool) {
	if intToCheck < 0 || intToCheck > 4294967295 {
		return uint32(0), false
	} else {
		return uint32(intToCheck), true
	}
}

func checkUint64(intToCheck int64) (uint64, bool) {
	if intToCheck < 0 {
		return uint64(0), false
	} else {
		return uint64(intToCheck), true
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
