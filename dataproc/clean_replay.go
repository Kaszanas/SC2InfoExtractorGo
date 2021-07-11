package dataproc

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"

	"github.com/Kaszanas/GoSC2Science/datastruct"
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/Kaszanas/GoSC2Science/settings"
)

func cleanReplay(replayData *rep.Rep, localizeMapsBool bool, localizedMapsMap map[string]interface{}) (bool, data.CleanedReplay) {

	log.Info("Entered cleanReplay()")

	// Restructure replay:
	structuredReplayData, redefOk := redifineReplayStructure(replayData, localizeMapsBool, localizedMapsMap)
	if !redefOk {
		log.Error("Error in redefining replay structure.")
		return false, data.CleanedReplay{}
	}

	// TODO: This needs to be controlled from outside of stringify_replay in case other users don't want to receive clean data.
	if !cleanUnusedGameEvents(&structuredReplayData) {
		log.Error("Error in cleaning the replay structure.")
		return false, data.CleanedReplay{}
	}

	return true, structuredReplayData
}

func cleanUnusedGameEvents(replayData *datastruct.CleanedReplay) bool {

	var cleanedGameEvents []s2prot.Struct
	for _, event := range replayData.GameEvents {
		if !contains(settings.UnusedGameEvents, event["evtTypeName"].(string)) {
			cleanedGameEvents = append(cleanedGameEvents, event)
		}
	}

	replayData.GameEvents = cleanedGameEvents

	return true
}
