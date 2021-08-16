package dataproc

import (
	"math"
	"strconv"
	"strings"

	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

const (
	Ranked1v1 = 1 << iota
	Ranked2v2
	Ranked3v3
	Ranked4v4
	RankedArchon
	Custom1v1
	Custom2v2
	Custom3v3
	Custom4v4
	CustomFFA
)

var gameModeList = []int{
	Ranked1v1,
	Ranked2v2,
	Ranked3v3,
	Ranked4v4,
	RankedArchon,
	Custom1v1,
	Custom2v2,
	Custom3v3,
	Custom4v4,
	CustomFFA,
}

// gameModeFiltersMapping contains information about different game modes and how to verify them.
var gameModeFiltersMapping = map[int]VerifyGameInfo{
	Ranked1v1:    {isAutoMatchMaking: true, maxPlayers: 2, isCompetitiveOrRanked: true},
	Ranked2v2:    {isAutoMatchMaking: true, maxPlayers: 4, isCompetitiveOrRanked: true},
	Ranked3v3:    {isAutoMatchMaking: true, maxPlayers: 6, isCompetitiveOrRanked: true},
	Ranked4v4:    {isAutoMatchMaking: true, maxPlayers: 8, isCompetitiveOrRanked: true},
	RankedArchon: {isAutoMatchMaking: true, maxPlayers: 4, isCompetitiveOrRanked: true},
	Custom1v1:    {isAutoMatchMaking: false, maxPlayers: 2, isCompetitiveOrRanked: false},
	Custom2v2:    {isAutoMatchMaking: false, maxPlayers: 4, isCompetitiveOrRanked: false},
	Custom3v3:    {isAutoMatchMaking: false, maxPlayers: 6, isCompetitiveOrRanked: false},
	Custom4v4:    {isAutoMatchMaking: false, maxPlayers: 8, isCompetitiveOrRanked: false},
	CustomFFA:    {isAutoMatchMaking: false, maxPlayers: 8, isCompetitiveOrRanked: false},
}

type VerifyGameInfo struct {
	isAutoMatchMaking     bool
	maxPlayers            int
	isCompetitiveOrRanked bool
}

// Integrity
// checkIntegrity verifies if the internal saved state of the replayData matches against structures with redundant information.
func checkIntegrity(replayData *rep.Rep) bool {

	log.Info("Entered checkIntegrity()")
	maxPlayers := replayData.InitData.GameDescription.MaxPlayers()
	replayDetails := replayData.Details

	// Technically there cannot be more than 15 human players!
	// Based on: https://s2editor-tutorials.readthedocs.io/en/master/01_Introduction/009_Player_Properties.html
	if maxPlayers > 16 || maxPlayers < 1 {
		log.WithField("maxPlayers", maxPlayers).Error("Integrity check failed! maxPlayers is not within the legal game engine range!")
		return false
	}

	// Map name of a replay is available in two places in the parsed data, if they mismatch then integrity test fails:
	if replayData.Metadata.Title() != replayDetails.Title() {
		log.WithFields(log.Fields{"metadataTitle": replayData.Metadata.Title(), "replayDetailsTitle": replayDetails.Title()}).Error("Integrity check failed! metadataTitle does not match replayDetailsTitle!")
		return false
	}

	// Checking if player list from replayDetails is of the same length as ToonPlayerDescMap:
	replayDetailsPlayerListLength := len(replayDetails.Players())
	toonPlayerDescMapLength := len(replayData.TrackerEvts.ToonPlayerDescMap)
	if replayDetailsPlayerListLength != toonPlayerDescMapLength {
		log.WithFields(log.Fields{"replayDetailsPlayerListLength": replayDetailsPlayerListLength, "toonPlayerDescMapLength": toonPlayerDescMapLength}).Error("Integrity check failed! length of players mismatch!")
		return false
	}

	metadatBaseBuildString := strings.Replace(replayData.Metadata.BaseBuild(), "Base", "", -1)
	metadataBaseBuildInt, err := strconv.Atoi(metadatBaseBuildString)
	if err != nil {
		log.Info("Integrity check failed! Cannot convert replayData.Metadata.BaseBuild() to integer!")
		return false
	}
	// Checking if game version contained in header fits the one that is in metadata:
	metadataBaseBuildInt64 := int64(metadataBaseBuildInt)
	headerBaseBuild := replayData.Header.BaseBuild()
	if headerBaseBuild != metadataBaseBuildInt64 {
		log.WithFields(log.Fields{"metadataBaseBuildInt64": metadataBaseBuildInt64, "headerBaseBuild": headerBaseBuild}).Error("Integrity check failed! Build version mismatch!")
		return false
	}

	gameOptions := replayData.InitData.GameDescription.GameOptions
	gameOptionsAmm := gameOptions.Amm()
	gameOptionsBattleNet := gameOptions.BattleNet()
	if gameOptionsAmm != gameOptionsBattleNet {
		log.WithFields(log.Fields{"gameOptionsAmm": gameOptionsAmm, "gameOptionsBattleNet": gameOptionsBattleNet}).Error("Integrity check failed! Amm and Battle.net mismatch")
		return false
	}

	log.Info("Integrity checks passed! Returning from checkIntegrity()")
	return true
}

// Validity
// validateReplay performs programmatically hardcoded checks in order to verify if the file is within "common sense" values.
func validateReplay(replayData *rep.Rep) bool {

	log.Info("Entered validateData()")
	playerList := replayData.Metadata.Players()

	if len(playerList) == 2 {
		absoluteMMRDifference := math.Abs(playerList[0].MMR() - playerList[1].MMR())
		// Around 1200 MMR:
		if absoluteMMRDifference > 1200 {
			log.Error("MMR Difference was found to be to big! validateData() failed, returning!")
			return false
		}
	}

	// In the history of StarCraft II there was no player that reached 8000 MMR so it should be below 8000 for all of the replays:
	for _, playerStats := range playerList {

		// Currently no player is 8000
		if playerStats.MMR() > 8000 {
			log.Error("Data validation failed! One of the players MMR is higher than 8000! Returning")
			return false
		}

		if playerStats.APM() == 0 {
			log.Error("Data validation failed! One of the players APM is equal to 0! Returning")
			return false
		}
	}

	log.Info("Finished validateData(), returning")
	return true
}

// Filtering
// checkGameMode performs the check against a HEX 0xFFFFFFFF getGameModeFlag to verify if the currently processed replay game mode is correct.
func checkGameMode(replayData *rep.Rep, getGameModeFlag int) bool {
	log.Info("Entered checkGameMode()")
	result := false

	for _, value := range gameModeList {

		if getGameModeFlag&value != 0 {
			result = result || checkGameParameters(replayData, gameModeFiltersMapping[value])
		}
	}

	log.Info("Finished checkGameMode()")
	return result
}

// checkGameParameters takes in a VerifyGameInfo struct that containts information about specific game mode filtering based on available data in the replay file:
func checkGameParameters(replayData *rep.Rep, gameInfoFilter VerifyGameInfo) bool {

	log.Info("Entered checkGameParameters()")

	if !checkNumberOfPlayers(replayData, gameInfoFilter.maxPlayers) {
		log.Error("Game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions
	gameOptionsAmm := gameOptions.Amm()

	if gameOptionsAmm != gameInfoFilter.isAutoMatchMaking {
		log.WithFields(log.Fields{"gameOptionsAmm": gameOptionsAmm, "isAutoMatchMaking": gameInfoFilter.isAutoMatchMaking}).Error("Game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	competitiveOrRanked := gameOptions.CompetitiveOrRanked()
	if competitiveOrRanked != gameInfoFilter.isCompetitiveOrRanked {
		log.WithFields(log.Fields{"competitiveOrRanked": competitiveOrRanked, "isCompetitiveOrRanked": gameInfoFilter.isCompetitiveOrRanked}).Error("Game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	maxPlayers := gameDescription.MaxPlayers()
	if maxPlayers != int64(gameInfoFilter.maxPlayers) {
		log.WithFields(log.Fields{"maxPlayers": maxPlayers, "gameInfoFilter.maxPlayers": gameInfoFilter.maxPlayers}).Error("Game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	log.Info("Finished checkGameParameters()")
	return true

}

// checkNumberOfPlayers verifies and checks if the number of players is correct for a given game mode.
func checkNumberOfPlayers(replayData *rep.Rep, requiredNumber int) bool {

	log.Info("Entered checkNumberOfPlayers()")

	playerList := replayData.Metadata.Players()
	numberOfPlayers := len(playerList)

	if numberOfPlayers != requiredNumber {
		log.WithFields(log.Fields{"len(playerList)": numberOfPlayers, "requiredNumber": requiredNumber}).Error("Integrity check failed number of players is not right!")
		return false
	}

	log.Info("Finished checkNumberOfPlayers(), returning")
	return true
}

// checkBlizzardMap verifies if the currently processed replay was played using a Blizzard official map.
func checkBlizzardMap(replayData *rep.Rep) bool {

	log.Info("Entered checkBlizzardMap()")

	gameDescIsBlizzardMap := replayData.InitData.GameDescription.IsBlizzardMap()
	detailsIsBlizzardMap := replayData.Details.IsBlizzardMap()

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if gameDescIsBlizzardMap == detailsIsBlizzardMap {
		log.Error("Integrity failed! isBlizzardMap information is inconsistent within a processed file!")
	}

	if !gameDescIsBlizzardMap {
		log.Error("Detected that the replay was played on a non-Blizzard map in gameDescription, returning")
		return false
	}

	if !detailsIsBlizzardMap {
		log.Error("Detected that the replay was played on a non-Blizzard map in gameDetails, returning")
		return false
	}

	log.Info("Finished checkBlizzardMap(), returning")
	return true
}
