package dataproc

import (
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
)

// TODO: Add Error handling, as currently there is absolutely no information about if the operations are correct or not.
func generateReplaySummary(replayData *data.CleanedReplay, summaryStruct *data.ReplaySummary) {

	// Game version histogram:
	var gameVersionFields = []string{"baseBuild", "build", "flags", "major", "minor", "revision"}

	// Getting the needed information out of the replay:
	replayHeader := replayData.Header

	// Check the map for every defined field and increment its value:
	for _, keyField := range gameVersionFields {
		key := strconv.FormatInt(replayHeader.Version[keyField].(int64), 10)
		// TODO: summaryStruct.Summary.GameVersions needs fields for every game version field that there is...
		keyExistsIncrementValue(key, summaryStruct.Summary.GameVersions)
	}

	replayMetadata := replayData.Metadata
	// GameDuration histogram:
	replayDuration := strconv.Itoa(int(replayMetadata.Duration.Seconds()))
	keyExistsIncrementValue(replayDuration, summaryStruct.Summary.GameTimes)

	// TODO: This needs to be checked for different language versions of the SC2 game.
	// This might require using the map checksums or some other additional information to synchronize.
	// MapsUsed histogram:
	replayMap := replayMetadata.MapName
	keyExistsIncrementValue(replayMap, summaryStruct.Summary.Maps)

	// Races used histogram:
	for _, player := range replayMetadata.Players {
		playerRace := player.AssignedRace
		keyExistsIncrementValue(playerRace, summaryStruct.Summary.Races)
	}

	// Dates of replays histogram:
	replayYear, replayMonth, replayDay := replayData.Details.TimeUTC.Date()
	dateString := strconv.Itoa(replayYear) + "-" + strconv.Itoa(int(replayMonth)) + "-" + strconv.Itoa(replayDay)
	keyExistsIncrementValue(dateString, summaryStruct.Summary.Dates)

	// Server information histogram. Region etc.

	// Amount of different units used (histogram of units used). Is this needed?

	// Histograms for maximum game time in different matchups. PvP, ZvP, TvP, ZvT, TvT, ZvZ

	// How many unique accounts were found

}

func keyExistsIncrementValue(key string, mapToCheck map[string]int64) {
	if val, ok := mapToCheck[key]; ok {
		mapToCheck[key] = val + 1
	} else {
		mapToCheck[key] = 1
	}
}
