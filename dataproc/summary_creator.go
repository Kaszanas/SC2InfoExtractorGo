package dataproc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
	log "github.com/sirupsen/logrus"
)

// generateReplaySummary accesses the data that is within cleaned replay
// and extracts information for visualization purposes.
func generateReplaySummary(
	replayData *replay_data.CleanedReplay,
	summaryStruct *persistent_data.ReplaySummary,
) {

	log.Debug("Entered generateReplaySummary()")

	// GameVersion information:
	var gameVersionString string
	gameVersionString = replayData.Metadata.GameVersion
	if gameVersionString == "" {
		// Accessing another data structure that holds game version string:
		gameVersionString = replayData.Header.Version
	}

	incrementIfKeyExists(gameVersionString, summaryStruct.Summary.GameVersions)
	log.Info("Finished incrementing replayData.Metadata.GameVersion")

	// REVIEW: This seems to be left as legacy:
	// replayMetadata := replayData.Metadata
	// GameDuration histogram:
	replayDuration := fmt.Sprintf("%f", float64(replayData.Header.ElapsedGameLoops)/22.4)

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

	// GameTimes per year histogram:
	// REVIEW: This seems to be left as legacy:
	// incrementNestedGameTimeIfKeyExists(strconv.Itoa(replayYear), replayDuration, summaryStruct.Summary.DatesGameTimes.GameTimes)
	// GameTimes per year-month histogram:
	incrementNestedGameTimeIfKeyExists(
		strconv.Itoa(replayYear)+"-"+strconv.Itoa(int(replayMonth)),
		replayDuration,
		summaryStruct.Summary.DatesGameTimes.GameTimes)
	// GameTimes per map histogram:
	incrementNestedGameTimeIfKeyExists(
		replayMap,
		replayDuration,
		summaryStruct.Summary.MapsGameTimes.GameTimes)

	// Server information histogram:
	// TODO: Verify if this can be accessed differently:
	singleLoop := false
	for _, player := range replayData.ToonPlayerDescMap {
		// This information is required only once per game:
		if !singleLoop {
			incrementIfKeyExists(player.Region, summaryStruct.Summary.Servers)
		}
		singleLoop = true
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

	log.Debug("Finished generateReplaySummary()")
}

// checkMatchup verifies the matchup string, increments the value of a counter
// for the matching matchup and returns a boolean that specifies if a matchup was matched.
func checkMatchupIncrementCount(
	matchupString string,
	summaryStruct *persistent_data.ReplaySummary,
	gameTimeString string) bool {

	log.Debug("Entered checkMatchup()")

	if matchupString == "TerrTerr" {
		log.Info("Found matchup to be TvT")
		incrementIfKeyExists("TvT", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(
			gameTimeString,
			summaryStruct.Summary.MatchupGameTimes.TvTMatchup)
		return true
	}
	if matchupString == "ProtProt" {
		log.Debug("Found matchup to be PvP")
		incrementIfKeyExists("PvP", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(
			gameTimeString,
			summaryStruct.Summary.MatchupGameTimes.PvPMatchup)
		return true
	}
	if matchupString == "ZergZerg" {
		log.Debug("Found matchup to be ZvZ")
		incrementIfKeyExists("ZvZ", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(
			gameTimeString,
			summaryStruct.Summary.MatchupGameTimes.ZvZMatchup)
		return true
	}
	if strings.Contains(matchupString, "Prot") && strings.Contains(matchupString, "Terr") {
		log.Debug("Found matchup to be PvT")
		incrementIfKeyExists("PvT", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(
			gameTimeString,
			summaryStruct.Summary.MatchupGameTimes.PvTMatchup)
		return true
	}
	if strings.ContainsAny(matchupString, "Zerg") && strings.Contains(matchupString, "Terr") {
		log.Debug("Found matchup to be ZvT")
		incrementIfKeyExists("ZvT", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(
			gameTimeString,
			summaryStruct.Summary.MatchupGameTimes.TvZMatchup)
		return true
	}
	if strings.ContainsAny(matchupString, "Zerg") && strings.Contains(matchupString, "Prot") {
		log.Debug("Found matchup to be ZvP")
		incrementIfKeyExists("ZvP", summaryStruct.Summary.MatchupCount)
		incrementIfKeyExists(
			gameTimeString,
			summaryStruct.Summary.MatchupGameTimes.PvZMatchup)
		return true
	}

	log.Info("Failed checkMatchup(), no matchup was found!")
	return false
}

// incrementIfKeyExists verifies if a key exists in a map and increments
// the value of a counter that is within a specific key.
func incrementIfKeyExists(key string, mapToCheck map[string]int64) {
	log.Debug("Entered keyExistsIncrementValue()")

	if val, ok := mapToCheck[key]; ok {
		mapToCheck[key] = val + 1
		log.Debug("Finished keyExistsIncrementValue(), value incremented")
	} else {
		mapToCheck[key] = 1
		log.Debug("Finished keyExistsIncrementValue(), new value added")
	}

	log.Debug("Finished keyExistsIncrementValue()")
}

func incrementNestedGameTimeIfKeyExists(
	key string,
	gameTime string,
	mapToCheck map[string]map[string]int64,
) {

	log.Debug("Entered incrementNestedGameTimeIfKeyExists()")

	if keyDateMap, ok := mapToCheck[key]; ok {
		if val, ok := keyDateMap[gameTime]; ok {
			keyDateMap[key] = val + 1
			log.Debug("Finished incrementNestedGameTimeIfKeyExists(), value incremented")
		} else {
			keyDateMap[key] = 1
			log.Debug("Finished incrementNestedGameTimeIfKeyExists(), new value added")
		}
	} else {
		mapToCheck[key] = map[string]int64{gameTime: 1}
	}

}
