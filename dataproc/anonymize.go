package dataproc

import (
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
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

	// Map toon to the nickname:
	var toonToNicknameMap map[string]string
	var newToonDescMap map[string]*rep.PlayerDesc
	// Iterate over players:
	for _, playerData := range replayData.Details.PlayerList {
		// Iterate over Toon description map:
		for toon, playerDesc := range replayData.ToonPlayerDescMap {
			// Checking if the SlotID and TeamID matches:
			if playerDesc.SlotID == int64(playerData.TeamID) {
				toonToNicknameMap[toon] = playerData.Name
				// Checking if the player toon was already anonymized (toons are unique, nicknames are not)
				anonymizedID, ok := persistPlayerNicknames[toon]
				if ok {
					playerData.Name = strconv.Itoa(anonymizedID)
					anonymizeToonDescMap(playerDesc, newToonDescMap, strconv.Itoa(anonymizedID))
				} else {
					persistPlayerNicknames[toon] = playerCounter
					anonymizeToonDescMap(playerDesc, newToonDescMap, strconv.Itoa(anonymizedID))
					playerCounter++
				}

				// TODO: Transfer the structure from playerDesc to a new *rep.PlayerDesc

				// TODO: Add the new *rep.PlayerDesc to the newToonDescMap

			}
		}
	}

	// TODO: After the loop completes replace the replayData.ToonDesc map with newToonDescMap.

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

func anonymizeToonDescMap(playerDesc *rep.PlayerDesc, toonDescMap map[string]*rep.PlayerDesc, anonymizedID string) {

	newPlayerDesc := rep.PlayerDesc{
		PlayerID:            playerDesc.PlayerID,
		SlotID:              playerDesc.SlotID,
		UserID:              playerDesc.UserID,
		StartLocX:           playerDesc.StartLocX,
		StartLocY:           playerDesc.StartLocY,
		StartDir:            playerDesc.StartDir,
		SQ:                  playerDesc.SQ,
		SupplyCappedPercent: playerDesc.SupplyCappedPercent,
	}

	toonDescMap[anonymizedID] = &newPlayerDesc

}
