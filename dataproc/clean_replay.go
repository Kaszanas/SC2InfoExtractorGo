package dataproc

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"

	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

// cleanReplay gathers functions that perform redefining "cleaning" of replay structure and cleans up events that are unused.
func cleanReplay(replayData *rep.Rep, localizeMapsBool bool, localizedMapsMap map[string]interface{}, performCleanupBool bool) (bool, data.CleanedReplay) {

	log.Info("Entered cleanReplay()")

	// Restructure replay:
	structuredReplayData, redefOk := redifineReplayStructure(replayData, localizeMapsBool, localizedMapsMap)
	if !redefOk {
		log.Error("Error in redefining replay structure.")
		return false, data.CleanedReplay{}
	}

	// Converting coordinates to fit the original map x, y ranges:
	if !convertCoordinates(&structuredReplayData) {
		log.Error("Error when converting coordinates.")
		return false, data.CleanedReplay{}
	}

	// Cleaning unused game events
	if performCleanupBool && !cleanUnusedGameEvents(&structuredReplayData) {
		log.Error("Error in cleaning the replay structure.")
		return false, data.CleanedReplay{}
	}

	log.Info("Finished cleanReplay()")
	return true, structuredReplayData
}

// cleanUnusedGameEvents checks against settings.UnusedGameEvents and creates new GameEvents structure without certain events.
func cleanUnusedGameEvents(replayData *data.CleanedReplay) bool {
	log.Info("Entered cleanUnusedGameEvents()")

	var cleanedGameEvents []s2prot.Struct
	for _, event := range replayData.GameEvents {
		if !contains(settings.UnusedGameEvents, event["evtTypeName"].(string)) {
			cleanedGameEvents = append(cleanedGameEvents, event)
		}
	}

	replayData.GameEvents = cleanedGameEvents

	log.Info("Finished cleanUnusedGameEvents()")
	return true
}
