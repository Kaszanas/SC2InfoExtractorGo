package dataproc

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
	log "github.com/sirupsen/logrus"
)

func anonimizeReplay(replayData *data.CleanedReplay) bool {

	if !anonimizeMessageEvents(replayData) {
		log.Error("Failed to anonimize messageEvents.")
		return false
	}
	if !anonymizePlayers(replayData) {
		log.Error("Failed to anonimize player information.")
		return false
	}

	return true
}

func anonymizePlayers(replayData *data.CleanedReplay) bool {

	// TODO: Introduce logging.
	// TODO: Name of the players should be anonymized to the same persistent value that the Toon will be anonymized.
	// Rhis means that the code should access the Toon information first and then replace respective information everywhere.

	// Nickname anonymization
	var persistPlayerNicknames map[string]int
	playerCounter := 0

	var toonToNicknameMap map[string]string
	for toon, player := range replayData.ToonPlayerDescMap {
		// Map toon to the nickname

		if player.PlayerID == 1 {
			toonToNicknameMap[toon] = "something"
		}

	}

	// Access the information that needs to be anonymized
	for _, player := range replayData.Details.PlayerList {
		// Check if it exists in some kind of persistent source that is used for the sake of anonymization
		anonymizedID, ok := persistPlayerNicknames[player.Name]
		if ok {
			// Replace the information within the original data structure with the persistent version from a variable or the file.
			player.Name = string(anonymizedID)
		} else {
			persistPlayerNicknames[player.Name] = playerCounter
			playerCounter++
		}
	}

	return true
}

func anonimizeMessageEvents(replayData *data.CleanedReplay) bool {

	var anonymizedMessageEvents []s2prot.Struct
	for _, event := range replayData.MessageEvents {
		eventType := event["evtTypeName"].(string)
		if !contains(settings.UnusedMessageEvents, eventType) {
			anonymizedMessageEvents = append(anonymizedMessageEvents, event)
		}
	}

	return true
}
