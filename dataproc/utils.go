package dataproc

// TODO: Introduce logging.
func checkClan(clanTag string) bool {
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

func checkUint8(intToCheck int) bool {

	if intToCheck < 0 || intToCheck > 255 {
		return false
	} else {
		return true
	}
}

func checkUint16(intToCheck int) bool {
	if intToCheck < 0 || intToCheck > 65535 {
		return false
	} else {
		return true
	}
}

func checkUint32(intToCheck int) bool {
	if intToCheck < 0 || intToCheck > 4294967295 {
		return false
	} else {
		return true
	}
}
