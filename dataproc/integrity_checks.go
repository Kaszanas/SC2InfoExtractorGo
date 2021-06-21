package dataproc

import (
	"math"

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

var gameModeFiltersMapping = map[int]VerifyGameInfo{
	Ranked1v1:    VerifyGameInfo{isAutoMatchMaking: true, maxPlayers: 2, isCompetitiveOrRanked: true},
	Ranked2v2:    VerifyGameInfo{isAutoMatchMaking: true, maxPlayers: 4, isCompetitiveOrRanked: true},
	Ranked3v3:    VerifyGameInfo{isAutoMatchMaking: true, maxPlayers: 6, isCompetitiveOrRanked: true},
	Ranked4v4:    VerifyGameInfo{isAutoMatchMaking: true, maxPlayers: 8, isCompetitiveOrRanked: true},
	RankedArchon: VerifyGameInfo{isAutoMatchMaking: true, maxPlayers: 4, isCompetitiveOrRanked: true},
	Custom1v1:    VerifyGameInfo{isAutoMatchMaking: false, maxPlayers: 2, isCompetitiveOrRanked: false},
	Custom2v2:    VerifyGameInfo{isAutoMatchMaking: false, maxPlayers: 4, isCompetitiveOrRanked: false},
	Custom3v3:    VerifyGameInfo{isAutoMatchMaking: false, maxPlayers: 6, isCompetitiveOrRanked: false},
	Custom4v4:    VerifyGameInfo{isAutoMatchMaking: false, maxPlayers: 8, isCompetitiveOrRanked: false},
	CustomFFA:    VerifyGameInfo{isAutoMatchMaking: false, maxPlayers: 8, isCompetitiveOrRanked: false},
}

type VerifyGameInfo struct {
	isAutoMatchMaking     bool
	maxPlayers            int
	isCompetitiveOrRanked bool
}

func checkGame(replayData *rep.Rep, getGameModeFlag int) bool {
	result := false

	for _, value := range gameModeList {

		if getGameModeFlag&value != 0 {
			result = result || checkGameParameters(replayData, gameModeFiltersMapping[value])
		}
	}

	return result
}

// checkGameParameters takes in a VerifyGameInfo struct that containts information about specific game mode filtering based on available data in the replay file:
func checkGameParameters(replayData *rep.Rep, gameInfo VerifyGameInfo) bool {

	if !checkNumberOfPlayers(replayData, gameInfo.maxPlayers) {
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != gameInfo.isAutoMatchMaking {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != gameInfo.isCompetitiveOrRanked {
		return false
	}

	if gameDescription.MaxPlayers() != int64(gameInfo.maxPlayers) {
		return false
	}

	return true

}

// Integrity
func checkIntegrity(replayData *rep.Rep, checkIntegrityBool bool, checkGameModeInt int) bool {

	return true
}

// Validity
func validateData(replayData *rep.Rep) bool {

	// Hand picked values for data validation of most probable data that can be met:

	// Check gameEvents "userOptions" "buildNum" and "baseBuildNum" against "header" information:
	playerList := replayData.Metadata.Players()

	if len(playerList) == 2 {
		absoluteMMRDifference := math.Abs(playerList[0].MMR() - playerList[1].MMR())
		// Around 1200 MMR:
		if absoluteMMRDifference > 1200 {
			log.Error("")
			return false
		}
	}

	// MMR should be below 8000 for all of the replays:
	for _, playerStats := range playerList {

		// Currently no player is 8000
		if playerStats.MMR() > 8000 {
			log.Error("Data validation failed! One of the players MMR is higher than 8000!")
			return false
		}

		if playerStats.APM() == 0 {
			log.Error("Data validation failed! One of the players APM is equal to 0!")
			return false
		}
	}

	return true
}

// Filtering
func checkGameMode(checkGameMode int) bool {

	return true

}

func checkNumberOfPlayers(replayData *rep.Rep, requiredNumber int) bool {
	// Check gameEvents "userOptions" "buildNum" and "baseBuildNum" against "header" information:
	playerList := replayData.Metadata.Players()

	if len(playerList) != requiredNumber {
		log.Error("Integrity check failed number of players is not right!")
		return false
	}
	return true
}

func checkBlizzardMap(replayData *rep.Rep) bool {

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

	return true
}
