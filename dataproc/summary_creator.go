package dataproc

import (
	"strconv"
	"strings"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	log "github.com/sirupsen/logrus"
)

// TODO: Add Error handling, as currently there is absolutely no information about if the operations are correct or not.
// generateReplaySummary accesses the data that is within cleaned replay and extracts information for visualization purposes.
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

	// Counting different units that were spawned in a game:
	for _, event := range replayData.TrackerEvents {
		// Counting the number of UnitBorn events to create histograms:
		eventType := event["evtTypeName"].(string)
		if eventType == "UnitBorn" {
			unitName := event["unitTypeName"].(string)
			keyExistsIncrementValue(unitName, summaryStruct.Summary.Units)
		}
	}
	log.Info("Finished incrementing summaryStruct.Summary.Units")

	// TODO: Histograms for maximum game time in different matchups. PvP, ZvP, TvP, ZvT, TvT, ZvZ

	// Creating matchup histograms:
	matchupString := replayData.Metadata.Players[0].AssignedRace + replayData.Metadata.Players[1].AssignedRace
	if !checkMatchup(matchupString, summaryStruct) {
		log.Error("Failed to increment matchup information!")
	}

	log.Info("Finished generateReplaySummary()")

}

// checkMatchup verifies the matchup string, increments the value of a counter of the matching matchup and returns a boolean that specifies if a matchup was matched.
func checkMatchup(matchupString string, summaryStruct *data.ReplaySummary) bool {
	log.Info("Entered checkMatchup()")

	if matchupString == "TerrTerr" {
		log.Info("Found matchup to be TvT")
		keyExistsIncrementValue("TvT", summaryStruct.Summary.MatchupHistograms)
		log.Info("Finished checkMatchup()")
		return true
	}
	if matchupString == "ProtProt" {
		log.Debug("Found matchup to be PvP")
		keyExistsIncrementValue("PvP", summaryStruct.Summary.MatchupHistograms)
		log.Info("Finished checkMatchup()")
		return true
	}
	if matchupString == "ZergZerg" {
		log.Debug("Found matchup to be ZvZ")
		keyExistsIncrementValue("ZvZ", summaryStruct.Summary.MatchupHistograms)
		log.Info("Finished checkMatchup()")
		return true
	}
	if strings.Contains(matchupString, "Prot") && strings.Contains(matchupString, "Terr") {
		log.Debug("Found matchup to be PvT")
		keyExistsIncrementValue("PvT", summaryStruct.Summary.MatchupHistograms)
		log.Info("Finished checkMatchup()")
		return true
	}
	if strings.ContainsAny(matchupString, "Zerg") && strings.Contains(matchupString, "Terr") {
		log.Debug("Found matchup to be ZvT")
		keyExistsIncrementValue("ZvT", summaryStruct.Summary.MatchupHistograms)
		log.Info("Finished checkMatchup()")
		return true
	}
	if strings.ContainsAny(matchupString, "Zerg") && strings.Contains(matchupString, "Prot") {
		log.Debug("Found matchup to be ZvP")
		keyExistsIncrementValue("ZvP", summaryStruct.Summary.MatchupHistograms)
		log.Info("Finished checkMatchup()")
		return true
	}

	log.Info("Failed checkMatchup(), no matchup was found!")
	return false
}

// keyExistsIncrementValue verifies if a key exists in a map and increments the value of a counter that is within a specific key.
func keyExistsIncrementValue(key string, mapToCheck map[string]int64) {
	log.Info("Entered keyExistsIncrementValue()")

	if val, ok := mapToCheck[key]; ok {
		mapToCheck[key] = val + 1
		log.Info("Finished keyExistsIncrementValue()")
	} else {
		mapToCheck[key] = 1
		log.Info("Finished keyExistsIncrementValue()")
	}
}
