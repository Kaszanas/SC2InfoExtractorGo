package dataproc

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
)

// TODO: Introduce logging.
// Helper function checking if a slice contains a string.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func anonymizePlayers() {

	// TODO: Anonymize the information about players.

	// Access the information that needs to be anonymized

	// Check if it exists in some kind of persistent source that is used for the sake of anonymization
	// This should be both performant and safe (can be a variable in memmory and a file on the drive that is written once every package)

	// Replace the information within the original data structure with the persistent version from a variable or the file.

}

// Anonymizes the replay and returns an error boolean
func anonymizeReplayData(replayData *data.CleanedReplay) bool {

	unusedMessageEvents := settings.UnusedMessageEvents()
	unusedGameEvents := settings.UnusedGameEvents()

	var anonymizedMessageEvents []s2prot.Event
	for _, event := range replayData.MessageEvents {
		if !contains(settings.UnusedMessageEvents(), event.Struct["evtTypeName"].(string)) {
			anonymizedMessageEvents = append(anonymizedMessageEvents, event)
		}
	}

	var anonymizedGameEvents []s2prot.Event
	for _, event := range replayData.GameEvents {
		if !contains(settings.UnusedGameEvents(), event.Struct["evtTypeName"].(string)) {
			anonymizedGameEvents = append(anonymizedGameEvents, event)
		}
	}

	replayData.MessageEvents = anonymizedMessageEvents
	replayData.GameEvents = anonymizedGameEvents

	return true
}
