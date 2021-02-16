package dataproc

// TODO: Introduce logging.
func checkClan(clanTag string) bool {
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

func checkUint8(intToCheck int64) (bool, uint8) {

	if intToCheck < 0 || intToCheck > 255 {
		return false, uint8(0)
	} else {
		return true, uint8(intToCheck)
	}
}

func checkUint16(intToCheck int64) (bool, uint16) {
	if intToCheck < 0 || intToCheck > 65535 {
		return false, uint16(0)
	} else {
		return true, uint16(intToCheck)
	}
}

func checkUint32(intToCheck int64) (bool, uint32) {
	if intToCheck < 0 || intToCheck > 4294967295 {
		return false, uint32(0)
	} else {
		return true, uint32(intToCheck)
	}
}

func checkUint64(intToCheck int64) (bool, uint64) {
	if intToCheck < 0 {
		return false, uint64(0)
	} else {
		return true, uint64(intToCheck)
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
