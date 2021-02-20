package dataproc

import (
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
)

func cleanReplay(replayData *rep.Rep) (bool, data.CleanedReplay) {

	log.Info("Entered cleanReplay()")
	successFlag := true

	// Restructure replay:
	structuredReplayData, redefOk := redifineReplayStructure(replayData)
	if !redefOk {
		log.Error("Error in redefining replay structure.")
		return !successFlag, data.CleanedReplay{}
	}

	// TODO: This needs to be controlled from outside of stringify_replay in case other users don't want to receive clean data.
	cleaningOk := cleanReplayStructure(&structuredReplayData)
	if !cleaningOk {
		log.Error("Error in cleaning the replay structure.")
		return !successFlag, data.CleanedReplay{}
	}

	return successFlag, structuredReplayData
}
