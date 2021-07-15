package dataproc

import (
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	log "github.com/sirupsen/logrus"
)

// TODO: Add Error handling, as currently there is absolutely no information about if the operations are correct or not.
func generateReplaySummary(replayData *data.CleanedReplay, summaryStruct *data.ReplaySummary) {

	log.Info("Entered generateReplaySummary()")

	// GameVersion information:
	keyExistsIncrementValue(replayData.Metadata.GameVersion, summaryStruct.Summary.GameVersions)
	log.Info("Finished incrementing replayData.Metadata.GameVersion")

	replayMetadata := replayData.Metadata
	// GameDuration histogram:
	replayDuration := strconv.Itoa(int(replayMetadata.Duration))
	keyExistsIncrementValue(replayDuration, summaryStruct.Summary.GameTimes)
	log.Info("Finished incrementing summaryStruct.Summary.GameTimes")

	// TODO: This needs to be checked for different language versions of the SC2 game.
	// This might require using the map checksums or some other additional information to synchronize.
	// MapsUsed histogram:
	replayMap := replayMetadata.MapName
	keyExistsIncrementValue(replayMap, summaryStruct.Summary.Maps)
	log.Info("Finished incrementing summaryStruct.Summary.Maps")

	// Races used histogram:
	for _, player := range replayMetadata.Players {
		playerRace := player.AssignedRace
		keyExistsIncrementValue(playerRace, summaryStruct.Summary.Races)
	}
	log.Info("Finished incrementing summaryStruct.Summary.Races")

	// Dates of replays histogram:
	replayYear, replayMonth, replayDay := replayData.Details.TimeUTC.Date()
	dateString := strconv.Itoa(replayYear) + "-" + strconv.Itoa(int(replayMonth)) + "-" + strconv.Itoa(replayDay)
	keyExistsIncrementValue(dateString, summaryStruct.Summary.Dates)
	log.Info("Finished incrementing summaryStruct.Summary.Dates")

	// Server information histogram:
	for _, player := range replayData.Details.PlayerList {
		keyExistsIncrementValue(player.Region, summaryStruct.Summary.Servers)
	}
	log.Info("Finished incrementing summaryStruct.Summary.Servers")

	// Amount of different units created (histogram of units used). Is this needed?
	// TODO: verify if this is needed it seems like too much information that is going to be generated:
	for _, event := range replayData.GameEvents {
		// TODO: Add another check not to include geisers, mineral fields and other unescessary information:
		if event["evtTypeName"].(string) == "UnitBorn" {
			keyExistsIncrementValue(event["unitTypeName"].(string), summaryStruct.Summary.Units)
		}
	}
	log.Info("Finished incrementing summaryStruct.Summary.Units")

	// // Histograms for maximum game time in different matchups. PvP, ZvP, TvP, ZvT, TvT, ZvZ

	// // TODO: flip the string to be the same always e.g. "TvP" == "PvT"
	// matchupString := replayData.Details.PlayerList[0].Race + "vs" + replayData.Details.PlayerList[1].Race

	// raceSlice := ["Zerg", "Protoss", ]

	// matchupSplit := strings.Split(matchupString, "Zerg")

	// keyExistsIncrementValue(matchupString, summaryStruct.Summary.MatchupHistograms)

	// // How many unique accounts were found:

}

func keyExistsIncrementValue(key string, mapToCheck map[string]int64) {
	if val, ok := mapToCheck[key]; ok {
		mapToCheck[key] = val + 1
	} else {
		mapToCheck[key] = 1
	}
}
