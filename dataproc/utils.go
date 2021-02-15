package dataproc

// TODO: Introduce logging.
func checkClan(clanTag string) bool {
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

func checkUint8(intToCheck int64) bool {

	if intToCheck < 0 || intToCheck > 255 {
		return false
	} else {
		return true
	}
}

func checkUint16(intToCheck int64) bool {
	if intToCheck < 0 || intToCheck > 65535 {
		return false
	} else {
		return true
	}
}

func checkUint32(intToCheck int64) bool {
	if intToCheck < 0 || intToCheck > 4294967295 {
		return false
	} else {
		return true
	}
}

func checkUint64(intToCheck int64) bool {
	if intToCheck < 0 {
		return false
	} else {
		return true
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
