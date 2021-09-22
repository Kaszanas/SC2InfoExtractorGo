package dataproc

import (
	"strconv"
	"strings"

	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
	log "github.com/sirupsen/logrus"
)

// generateReplaySummary accesses the data that is within cleaned replay and extracts information for visualization purposes.
func generateReplaySummary(replayData *data.CleanedReplay, summaryStruct *data.ReplaySummary) {

	log.Info("Entered generateReplaySummary()")

	// GameVersion information:
	gameVersionString := replayData.Metadata.GameVersion
	if gameVersionString == "" {
		// Accessing another data structure that holds game version string:
		gameVersionString = replayData.Header.Version.String()
	}

	incrementIfKeyExists(gameVersionString, summaryStruct.Summary.GameVersions)
	log.Info("Finished incrementing replayData.Metadata.GameVersion")

	replayMetadata := replayData.Metadata
	// GameDuration histogram:
	replayDuration := strconv.Itoa(int(replayMetadata.Duration))
	// If the game duration from metadata doesn't exist use the one from Header:
	if replayDuration == "" {
		replayDuration = strconv.Itoa(int(replayData.Header.DurationSeconds))
	}
	incrementIfKeyExists(replayDuration, summaryStruct.Summary.GameTimes)
	log.Info("Finished incrementing summaryStruct.Summary.GameTimes")

	// MapsUsed histogram:
	replayMap := replayData.Metadata.MapName
	incrementIfKeyExists(replayMap, summaryStruct.Summary.Maps)
	log.Info("Finished incrementing summaryStruct.Summary.Maps")

	// Races used histogram:
	for _, player := range replayData.ToonPlayerDescMap {
		playerRace := player.AssignedRace
		incrementIfKeyExists(playerRace, summaryStruct.Summary.Races)
	}
	log.Info("Finished incrementing summaryStruct.Summary.Races")

	// Dates of replays histogram:
	replayYear, replayMonth, replayDay := replayData.Details.TimeUTC.Date()
	dateString := strconv.Itoa(replayYear) + "-" + strconv.Itoa(int(replayMonth)) + "-" + strconv.Itoa(replayDay)
	incrementIfKeyExists(dateString, summaryStruct.Summary.Dates)
	log.Info("Finished incrementing summaryStruct.Summary.Dates")

	// Server information histogram:
	for _, player := range replayData.ToonPlayerDescMap {
		incrementIfKeyExists(player.Region, summaryStruct.Summary.Servers)
	}
	log.Info("Finished incrementing summaryStruct.Summary.Servers")

	// Counting different units that were spawned in a game:
	for _, event := range replayData.TrackerEvents {
		// Counting the number of UnitBorn events to create histograms:
		eventType := event["evtTypeName"].(string)
		if eventType == "UnitBorn" {

			// If the unit is not recognized as player controllable unit it is put to OtherUnits
			unitName := event["unitTypeName"].(string)
			if contains(settings.ExcludeUnitsFromSummary, unitName) {
				incrementIfKeyExists(unitName, summaryStruct.Summary.OtherUnits)
				continue
			}
			// If GhostAlternate -> change to Ghost?
			incrementIfKeyExists(unitName, summaryStruct.Summary.Units)
		}
	}
	log.Info("Finished incrementing summaryStruct.Summary.Units")

	var matchupString string
	for _, player := range replayData.ToonPlayerDescMap {
		matchupString += player.AssignedRace
	}
	// Incrementing both the count of matchup and the game time that the matchup had:
	if !checkMatchupIncrementCount(matchupString, summaryStruct, replayDuration) {
		log.Error("Failed to increment matchup information!")
	}

	log.Info("Finished generateReplaySummary()")

}

// checkMatchup verifies the matchup string, increments the value of a counter of the matching matchup and returns a boolean that specifies if a matchup was matched.
func checkMatchupIncrementCount(matchupString string, summaryStruct *data.ReplaySummary, gameTimeString string) bool {
	log.Info("Entered checkMatchup()")

	if matchupString == "TerrTerr" {
		log.Info("Found matchup to be TvT")
		incrementIfKeyExists("TvT", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(gameTimeString, summaryStruct.Summary.MatchupGameTimes.TvTMatchup)
		return true
	}
	if matchupString == "ProtProt" {
		log.Debug("Found matchup to be PvP")
		incrementIfKeyExists("PvP", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(gameTimeString, summaryStruct.Summary.MatchupGameTimes.PvPMatchup)
		return true
	}
	if matchupString == "ZergZerg" {
		log.Debug("Found matchup to be ZvZ")
		incrementIfKeyExists("ZvZ", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(gameTimeString, summaryStruct.Summary.MatchupGameTimes.ZvZMatchup)
		return true
	}
	if strings.Contains(matchupString, "Prot") && strings.Contains(matchupString, "Terr") {
		log.Debug("Found matchup to be PvT")
		incrementIfKeyExists("PvT", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(gameTimeString, summaryStruct.Summary.MatchupGameTimes.PvTMatchup)
		return true
	}
	if strings.ContainsAny(matchupString, "Zerg") && strings.Contains(matchupString, "Terr") {
		log.Debug("Found matchup to be ZvT")
		incrementIfKeyExists("ZvT", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(gameTimeString, summaryStruct.Summary.MatchupGameTimes.TvZMatchup)
		return true
	}
	if strings.ContainsAny(matchupString, "Zerg") && strings.Contains(matchupString, "Prot") {
		log.Debug("Found matchup to be ZvP")
		incrementIfKeyExists("ZvP", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(gameTimeString, summaryStruct.Summary.MatchupGameTimes.PvZMatchup)
		return true
	}

	log.Info("Failed checkMatchup(), no matchup was found!")
	return false
}

// incrementIfKeyExists verifies if a key exists in a map and increments the value of a counter that is within a specific key.
func incrementIfKeyExists(key string, mapToCheck map[string]int64) {
	log.Info("Entered keyExistsIncrementValue()")

	if val, ok := mapToCheck[key]; ok {
		mapToCheck[key] = val + 1
		log.Info("Finished keyExistsIncrementValue(), value incremented")
	} else {
		mapToCheck[key] = 1
		log.Info("Finished keyExistsIncrementValue(), new value added")
	}
}
