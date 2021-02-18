package dataproc

import "github.com/icza/s2prot/rep"

func checkIntegrity(replayData *rep.Rep) bool {

	// TODO: Check for every doubled information if it is the same with existing s2prot.Rep structures for data integrity validation.

	var checkSlice []bool

	// Checking if isBlizzardMap is the same in both of the available places:
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		checkSlice = append(checkSlice, true)
	}

	// Check gameEvents "userOptions" "buildNum" and "baseBuildNum" against "header" information:

	return true
}
