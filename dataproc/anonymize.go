package dataproc

import (
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func anonymizeReplay(replayData *data.CleanedReplay) bool {

	log.Info("Entered anonymizeReplay()")

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

	log.Info("Entererd anonymizePlayers().")
	// TODO: Introduce logging.
	// Nickname anonymization
	var persistPlayerNicknames map[string]int
	playerCounter := 0
	// Map toon to the nickname:
	var toonToNicknameMap map[string]string
	var newToonDescMap map[string]*rep.PlayerDesc

	var listOfStructs = make([]rep.PlayerDesc, 2)

	// Iterate over players:
	log.Info("Starting to iterate over replayData.Details.PlayerList.")
	for index, playerData := range replayData.Details.PlayerList {
		// Iterate over Toon description map:
		for toon, playerDesc := range replayData.ToonPlayerDescMap {
			// Checking if the SlotID and TeamID matches:
			if playerDesc.SlotID == int64(playerData.TeamID) {
				toonToNicknameMap[toon] = playerData.Name
				// Checking if the player toon was already anonymized (toons are unique, nicknames are not)
				anonymizedID, ok := persistPlayerNicknames[toon]
				if ok {
					// TODO: Add all of the other information that needs to be anonymized about the players:
					playerData.Name = strconv.Itoa(anonymizedID)
					anonymizeToonDescMap(&listOfStructs[index], newToonDescMap, strconv.Itoa(anonymizedID))
				} else {
					persistPlayerNicknames[toon] = playerCounter
					anonymizeToonDescMap(&listOfStructs[index], newToonDescMap, strconv.Itoa(anonymizedID))
					playerCounter++
				}
			}
		}
	}

	// Replacing Toon desc map with anonymmized version containing a persistent anonymized ID of the player:
	log.Info("Replacing ToonPlayerDescMap with anonymized version.")
	replayData.ToonPlayerDescMap = newToonDescMap

	return true
}

func anonimizeMessageEvents(replayData *data.CleanedReplay) bool {

	log.Info("Entered anonimizeMessageEvents().")
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

	log.Info("Entered anonymizeToonDescMap().")

	// Define new rep.PlayerDesc with old
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

	// Adding the new PlayerDesc
	log.Info("Adding new PlayerDesc to toonDescMap")
	toonDescMap[anonymizedID] = &newPlayerDesc

}
