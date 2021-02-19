package dataproc

import (
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
)

func generateSummary(replayData data.CleanedReplay, summaryInfo *data.PackageSummary) {

	// Game version histogram:
	var gameVersionFields = []string{"baseBuild", "build", "flags", "major", "minor", "revision"}

	// Getting the needed information out of the replay:
	replayHeader := replayData.Header

	// Check the map for every defined field and increment its value:
	for _, keyField := range gameVersionFields {
		key := strconv.FormatInt(replayHeader.Version[keyField].(int64), 10)
		summaryInfo.GameVersions = keyExistsIncrementValue(key, summaryInfo.GameVersions)
	}

	// GameDuration histogram:
	replayDuration := replayData.Metadata.Duration.String()
	summaryInfo.GameTimes = keyExistsIncrementValue(replayDuration, summaryInfo.GameTimes)

	// Game time histogram
	// Access the game duration

	// Maps used histogram (This needs to take into consideration that the maps might be named differently depending on what language version of the game was used?)
	// This might require using the map checksums or some other additional information to synchronize.

	// Race summary (This will be calculated on a replay by replay basis)

	// Amount of different units used (histogram of units used). Is this needed?

	// Dates of the replay when was the first recorded replay in the package when was the last recorded replay in the package.

	// Server information histogram. Region etc.

	// Histograms for maximum game time in different matchups. PvP, ZvP, TvP, ZvT, TvT, ZvZ

	// How many unique accounts were found

}

func keyExistsIncrementValue(key string, mapToCheck map[string]int64) map[string]int64 {
	if val, ok := mapToCheck[key]; ok {
		mapToCheck[key] = val + 1
	} else {
		mapToCheck[key] = 1
	}
	return mapToCheck
}
