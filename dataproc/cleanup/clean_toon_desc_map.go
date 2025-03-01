package cleanup

import (
	"fmt"
	"strings"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanToonDescMap copies the toon description map,
// changes the structure into a more readable form.
func CleanToonDescMap(
	replayData *rep.Rep,
	cleanedUserInitDataList []replay_data.CleanedUserInitData,
) (map[string]replay_data.EnhancedToonDescMap, bool) {

	dirtyToonPlayerDescMap := replayData.TrackerEvts.ToonPlayerDescMap

	// Merging data-structures to data.EnhancedToonDescMap
	enhancedToonDescMap := make(map[string]replay_data.EnhancedToonDescMap)
	for toonKey, playerDescription := range dirtyToonPlayerDescMap {
		// Solved https://github.com/Kaszanas/SC2InfoExtractorGo/issues/51
		// Initializing enhanced map from the dirtyToonPlayerDescMap:
		initializedToonDescMap := replay_data.EnhancedToonDescMap{
			PlayerID:            playerDescription.PlayerID,
			UserID:              playerDescription.UserID,
			SQ:                  playerDescription.SQ,
			SupplyCappedPercent: playerDescription.SupplyCappedPercent,
			StartDir:            playerDescription.StartDir,
			StartLocX:           playerDescription.StartLocX,
			StartLocY:           playerDescription.StartLocY,
		}
		enhancedToonDescMap[toonKey] = initializedToonDescMap

		// Merging information held in metadata.Players into data.EnhancedToonDescMap
		enhancedToonDescMap[toonKey] = mergeToonDescMapWithMetadata(
			replayData,
			playerDescription,
			enhancedToonDescMap[toonKey],
		)

		// Merging information contained in the details part of the replay:
		var err error
		enhancedToonDescMap[toonKey], err = mergeToonDescMapWithDetails(
			replayData,
			toonKey,
			enhancedToonDescMap[toonKey],
		)
		if err != nil {
			log.WithField("error", err.Error()).
				Error("Failed to merge toon desc map with details")
			return enhancedToonDescMap, false
		}

		// Merging information contained in the cleanedUserInitDataList:
		enhancedToonDescMap[toonKey], err = mergeToonDescMapWithInitPlayerList(
			cleanedUserInitDataList,
			enhancedToonDescMap[toonKey],
		)
		if err != nil {
			log.WithField("error", err.Error()).
				Error("Failed to merge toon desc map with init player list")
			return enhancedToonDescMap, false
		}

	}

	return enhancedToonDescMap, true
}

// mergeToonDescMapWithMetadata merges the rep.PlayerDesc with the game Metadata
// into the EnhancedToonDescMap
func mergeToonDescMapWithMetadata(
	replayData *rep.Rep,
	playerDescription *rep.PlayerDesc,
	enhancedToonDescMap replay_data.EnhancedToonDescMap,
) replay_data.EnhancedToonDescMap {

	metadataPlayers := replayData.Metadata.Players()
	if len(metadataPlayers) == 0 {
		log.Warn("No players found in metadata!")
		return enhancedToonDescMap
	}
	metadataEnhancedToonDescMap := enhancedToonDescMap

	for _, metadataPlayer := range metadataPlayers {

		metadataPlayerID := metadataPlayer.PlayerID()
		playerDescriptionPlayerID := playerDescription.PlayerID

		if metadataPlayerID != playerDescriptionPlayerID {
			continue
		}

		// Filling out struct fields:
		// What should be done in case if some of these fields are empty?
		metadataEnhancedToonDescMap.AssignedRace = metadataPlayer.AssignedRace()
		metadataEnhancedToonDescMap.SelectedRace = metadataPlayer.SelectedRace()
		metadataEnhancedToonDescMap.APM = metadataPlayer.APM()
		metadataEnhancedToonDescMap.MMR = metadataPlayer.MMR()
		metadataEnhancedToonDescMap.Result = metadataPlayer.Result()
	}
	return metadataEnhancedToonDescMap
}

// mergeToonDescMapWithDetails merges the rep.PlayerDesc with the game Details,
// into the EnhancedToonDescMap, and returns the EnhancedToonDescMap.
func mergeToonDescMapWithDetails(
	replayData *rep.Rep,
	toonKey string,
	enhancedToonDescMap replay_data.EnhancedToonDescMap,
) (replay_data.EnhancedToonDescMap, error) {

	detailsPlayers := replayData.Details.Players()

	if len(detailsPlayers) == 0 {
		log.Warn("No players found in details!")
		return enhancedToonDescMap, nil
	}

	detailsEnhancedToonDescMap := enhancedToonDescMap
	for _, player := range detailsPlayers {
		toonString := player.Toon.String()

		// Toon string doesn't match, keep looking for the right player:
		if toonString != toonKey {
			continue
		}

		// Found the right player, fill out the fields:
		detailsEnhancedToonDescMap.Name = player.Name

		// Checking if previously ran loop populated the Race information
		// REVIEW: Should this be a full race name?
		if detailsEnhancedToonDescMap.AssignedRace == "" {
			raceLetter := player.Race().Letter
			if raceLetter == 'T' {
				detailsEnhancedToonDescMap.AssignedRace = "Terr"
			}
			if raceLetter == 'P' {
				detailsEnhancedToonDescMap.AssignedRace = "Prot"
			}
			if raceLetter == 'Z' {
				detailsEnhancedToonDescMap.AssignedRace = "Zerg"
			}
		}

		resultMapPlayerResult := map[string]string{
			"Unknown": "Undecided",
			"Victory": "Win",
			"Defeat":  "Loss",
			"Tie":     "Tie",
		}

		playerResult := player.Result().String()
		playerResult = resultMapPlayerResult[playerResult]

		if detailsEnhancedToonDescMap.Result != "" {
			if detailsEnhancedToonDescMap.Result != playerResult {
				log.Warn("Player results are different!")
				return detailsEnhancedToonDescMap, fmt.Errorf("player results are different")
			}
		}

		// Result was empty, fill it out:
		if detailsEnhancedToonDescMap.Result == "" {
			detailsEnhancedToonDescMap.Result = playerResult
		}

		// Filling out struct fields:
		detailsEnhancedToonDescMap.Region = player.Toon.Region().Name
		detailsEnhancedToonDescMap.Realm = player.Toon.Realm().Name
		detailsEnhancedToonDescMap.Color.A = player.Color[0]
		detailsEnhancedToonDescMap.Color.B = player.Color[1]
		detailsEnhancedToonDescMap.Color.G = player.Color[2]
		detailsEnhancedToonDescMap.Color.R = player.Color[3]
		detailsEnhancedToonDescMap.Handicap = player.Handicap()
		// There should be only one player that fits the unique toon key,
		// if the player was found and the values were filled out we can exit the loop:
		break
	}

	return detailsEnhancedToonDescMap, nil
}

func mergeToonDescMapWithInitPlayerList(
	cleanedUserInitDataList []replay_data.CleanedUserInitData,
	enhancedToonDescMap replay_data.EnhancedToonDescMap,
) (replay_data.EnhancedToonDescMap, error) {

	if len(cleanedUserInitDataList) == 0 {
		log.Warn("No players found in cleanedUserInitDataList!")
		return enhancedToonDescMap, nil
	}

	// Merging cleanedUserInitDataList information into data.EnhancedToonDescMap:
	initEnhancedToonDescMap := enhancedToonDescMap
	for _, initPlayer := range cleanedUserInitDataList {

		toonMapName := initEnhancedToonDescMap.Name
		initPlayerName := initPlayer.Name

		// The names don't align, keep looking:
		if !strings.HasSuffix(toonMapName, initPlayerName) {
			continue
		}

		// Both names are empty, this cannot be correct:
		if initEnhancedToonDescMap.Name == "" && initPlayer.Name == "" {
			log.Error("Both player names are empty in mergeToonDescMapWithInitPlayerList!")
			return initEnhancedToonDescMap, fmt.Errorf("both player names are empty")
		}

		// If the name is empty in the EnhancedToonDescMap, fill it out from the initPlayer:
		if initEnhancedToonDescMap.Name == "" && initPlayer.Name != "" {
			initEnhancedToonDescMap.Name = initPlayer.Name
		}
		initEnhancedToonDescMap.HighestLeague = initPlayer.HighestLeague
		initEnhancedToonDescMap.IsInClan = initPlayer.IsInClan
		initEnhancedToonDescMap.ClanTag = initPlayer.ClanTag

		// No need to look any further:
		break
	}

	return initEnhancedToonDescMap, nil

}
