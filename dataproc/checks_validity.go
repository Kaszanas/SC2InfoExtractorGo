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

// gameModeFiltersMapping contains information about different game modes and how to verify them.
var gameModeFiltersMapping = map[int]VerifyGameInfo{
	Ranked1v1: {isAutoMatchMaking: true, maxPlayers: 2, isCompetitiveOrRanked: true},
	Ranked2v2: {isAutoMatchMaking: true, maxPlayers: 4, isCompetitiveOrRanked: true},
	Ranked3v3: {isAutoMatchMaking: true, maxPlayers: 6, isCompetitiveOrRanked: true},
	Ranked4v4: {isAutoMatchMaking: true, maxPlayers: 8, isCompetitiveOrRanked: true},
	// RankedArchon: {isAutoMatchMaking: true, maxPlayers: 4, isCompetitiveOrRanked: true},
	Custom1v1: {isAutoMatchMaking: false, maxPlayers: 2, isCompetitiveOrRanked: false},
	Custom2v2: {isAutoMatchMaking: false, maxPlayers: 4, isCompetitiveOrRanked: false},
	Custom3v3: {isAutoMatchMaking: false, maxPlayers: 6, isCompetitiveOrRanked: false},
	Custom4v4: {isAutoMatchMaking: false, maxPlayers: 8, isCompetitiveOrRanked: false},
	// CustomFFA: {isAutoMatchMaking: false, maxPlayers: 8, isCompetitiveOrRanked: false},
}

type VerifyGameInfo struct {
	isAutoMatchMaking     bool
	maxPlayers            int
	isCompetitiveOrRanked bool
}

// Validity
// validateReplay performs programmatically hardcoded checks in order to verify if the file is within "common sense" values.
func validate1v1Replay(replayData *rep.Rep) bool {

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

// checkBlizzardMap verifies if the currently processed replay was played using a Blizzard official map.
func checkBlizzardMap(replayData *rep.Rep) bool {

	log.Info("Entered checkBlizzardMap()")

	gameDescIsBlizzardMap := replayData.InitData.GameDescription.IsBlizzardMap()
	detailsIsBlizzardMap := replayData.Details.IsBlizzardMap()

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
