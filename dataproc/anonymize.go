package dataproc

import (
	"fmt"
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func anonymizeReplay(replayData *data.CleanedReplay, playersAnonymized map[string]int) bool {

	log.Info("Entered anonymizeReplay()")

	if !anonimizeMessageEvents(replayData) {
		log.Error("Failed to anonimize messageEvents.")
		return false
	}

	if !anonymizePlayers(replayData, playersAnonymized) {
		log.Error("Failed to anonimize player information.")
		return false
	}

	return true
}

func anonymizePlayers(replayData *data.CleanedReplay, playersAnonymized map[string]int) bool {

	log.Info("Entererd anonymizePlayers().")
	playerCounter := 0

	// TODO: if this is going to be persistent it needs to be passed as a pointer and live in the main function.

	var newToonDescMap = make(map[string]*rep.PlayerDesc)
	// var listOfStructs = make([]rep.PlayerDesc, 2)

	// Iterate over players:
	log.Info("Starting to iterate over replayData.Details.PlayerList.")
	// TODO: Something does not work here!!! Empty toon map is generated and stringified.
	for index, playerData := range replayData.Details.PlayerList {
		// Iterate over Toon description map:
		for toon, playerDesc := range replayData.ToonPlayerDescMap {
			// Checking if the SlotID and TeamID matches:
			if playerDesc.SlotID == int64(playerData.TeamID) {
				// Checking if the player toon was already anonymized (toons are unique, nicknames are not)
				// TODO: This line of code needs to use external file which will be updated per package.
				// TODO: The software should allow restarting processing from a package that errored out.
				anonymizedID, ok := persistPlayerNicknames[toon]
				if ok {
					// TODO: Add all of the other information that needs to be anonymized about the players:
					// Nickname anonymization:
					stringAnonymizedID := strconv.Itoa(anonymizedID)
					replayData.Details.PlayerList[index].Name = stringAnonymizedID
					// Toon anonymization:
					anonymizeToonDescMap(playerDesc, newToonDescMap, stringAnonymizedID, rep.PlayerDesc{})
				} else {
					// The toon was not ine the persistent map, add it:
					persistPlayerNicknames[toon] = playerCounter

					// Convert player counter to string to be used as new toon in the final map:
					stringAnonymizedID := strconv.Itoa(playerCounter)
					replayData.Details.PlayerList[index].Name = stringAnonymizedID

					anonymizeToonDescMap(playerDesc, newToonDescMap, stringAnonymizedID, rep.PlayerDesc{})
					playerCounter++
				}
			}
		}
	}

	// Replacing Toon desc map with anonymmized version containing a persistent anonymized ID of the player:
	log.Info("Replacing ToonPlayerDescMap with anonymized version.")
	replayData.ToonPlayerDescMap = newToonDescMap

	fmt.Println(replayData.ToonPlayerDescMap)

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

	replayData.MessageEvents = anonymizedMessageEvents

	return true
}

func anonymizeToonDescMap(playerDesc *rep.PlayerDesc, toonDescMap map[string]*rep.PlayerDesc, anonymizedID string, emptyPlayerDesc rep.PlayerDesc) {

	log.Info("Entered anonymizeToonDescMap().")

	// Define new rep.PlayerDesc with old
	emptyPlayerDesc = rep.PlayerDesc{
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
	toonDescMap[anonymizedID] = &emptyPlayerDesc

}
