package dataproc

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

// cleanReplay gathers functions that perform redefining
// "cleaning" of replay structure and cleans up events that are unused.
func extractReplayData(
	replayData *rep.Rep,
	englishToForeignMapping map[string]string,
	performCleanupBool bool,
) (bool, replay_data.CleanedReplay) {

	log.Debug("Entered cleanReplay()")

	// Restructure replay:
	structuredReplayData, redefOk := redifineReplayStructure(
		replayData,
		englishToForeignMapping,
	)
	if !redefOk {
		log.Error("Error in redefining replay structure.")
		return false, replay_data.CleanedReplay{}
	}

	// Cleaning unused message and game events
	if performCleanupBool {
		if !cleanUnusedMessageEvents(&structuredReplayData) {
			log.Error("Error in cleaning the message events.")
			return false, replay_data.CleanedReplay{}
		}
		if !cleanUnusedGameEvents(&structuredReplayData) {
			log.Error("Error in cleaning the game events.")
			return false, replay_data.CleanedReplay{}
		}
	}

	log.Debug("Finished cleanReplay()")
	return true, structuredReplayData
}

// cleanUnusedMessageEvents iterates over the message events and creates
// a new structure without the events that were hardcoded as redundant.
func cleanUnusedMessageEvents(replayData *replay_data.CleanedReplay) bool {

	log.Debug("Entered cleanUnusedMessageEvents()")

	var cleanMessageEvents []s2prot.Struct
	for _, event := range replayData.MessageEvents {
		if !contains(settings.UnusedMessageEvents, event["evtTypeName"].(string)) {
			cleanMessageEvents = append(cleanMessageEvents, event)
		}
	}

	replayData.MessageEvents = cleanMessageEvents

	log.Debug("Finished cleanUnusedMessageEvents()")
	return true
}

// cleanUnusedGameEvents checks against settings.UnusedGameEvents and
// creates a new GameEvents structure without certain events.
func cleanUnusedGameEvents(replayData *replay_data.CleanedReplay) bool {
	log.Debug("Entered cleanUnusedGameEvents()")

	var cleanedGameEvents []map[string]any
	for _, event := range replayData.GameEvents {

		if eventType, ok := event["evtTypeName"]; ok {
			if eventType == nil {
				log.Error("Failed to get evtTypeName from GameEvents, cannot check for unused events.")
				continue
			}

			castedEventType := eventType.(string)

			if !contains(settings.UnusedGameEvents, castedEventType) {
				cleanedGameEvents = append(cleanedGameEvents, event)
			}
		}

	}

	replayData.GameEvents = cleanedGameEvents

	log.Debug("Finished cleanUnusedGameEvents()")
	return true
}
