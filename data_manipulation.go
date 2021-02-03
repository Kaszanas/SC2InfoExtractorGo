package main

import "github.com/icza/s2prot/rep"

// type CleanedReplay struct {
// 	header rep.Header
// }

func deleteUnusedObjects(replayData *rep.Rep) *rep.Rep {

	// TODO: Copy types from icza package and create custom marshal that will not include fields which are not relevant to the dataset.

	// TODO: Clear the objects that will not be used in the final JSON

	return replayData
}

func anonymizeReplayData(replayData *rep.Rep) *rep.Rep {

	// TODO: Anonymize the information about players and chat events

	return replayData
}
