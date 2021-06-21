package dataproc

import (
	"math"

	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

type GameMode int

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

func (w GameMode) String() string {
	return [...]string{"AllGameModes", "Ranked1v1", "Ranked2v2", "Ranked3v3", "Ranked4v4", "RankedArchon", "Custom1v1", "Custom2v2", "Custom3v3", "Custom4v4", "CustomFFA"}[w-1]
}

func (w GameMode) EnumIndex() int {
	return int(w)
}

// TODO: Finish this
var gameModeFiltersMapping = map[int]VerifyGameInfo{
	Ranked1v1: VerifyGameInfo{isAutoMatchMaking: true, maxPlayers: 2, isCompetitiveOrRanked: true},
	Ranked2v2: VerifyGameInfo{}}

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

// Validity

// Filtering

func checkIntegrity(replayData *rep.Rep, checkIntegrityBool bool, checkGameModeInt int) bool {

	if checkGameModeInt == AllGameModes.EnumIndex() {
		log.Info("")
		if checkIntegrityBool {
			basicIntegrityOk := checkBasicIntegrity(replayData)
			if !basicIntegrityOk {
				log.Info("")
				return false
			}
		}
		log.Info("")
		return true
	}

	if checkGameModeInt == Ranked1v1.EnumIndex() {

		if !checkGameMode(checkGameModeInt) {
			return false
		}

		if checkIntegrityBool {
			basicIntegrityOk := checkBasicIntegrity(replayData)
			if !basicIntegrityOk {
				log.Info("")
				return false
			}
		}
	}

	if checkGameModeInt == Ranked2v2.EnumIndex() {
		if !checkGameMode(checkGameModeInt) {
			return false
		}
	}

	return true
}

func checkBasicIntegrity(replayData *rep.Rep) bool {

}

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
}

func checkGameMode(checkGameMode int) bool {

	log.Info("")
	is1v1RankedGameMode := checkRanked1v1(replayData)
	if !is1v1RankedGameMode {
		log.Info("")
		return false
	}

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

}

func checkRanked1v1(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 2) {
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != true {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != true {
		return false
	}

	if gameDescription.MaxPlayers() > 2 {
		return false
	}

	return true
}

func checkRanked2v2(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 4) {
		return false
	}

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		log.Error("Integrity failed! Map was found not to be a blizzard map!")
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != true {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != true {
		return false
	}

	if gameDescription.MaxPlayers() != 4 {
		return false
	}

	return true
}

func checkRanked3v3(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 6) {
		return false
	}

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		log.Error("Integrity failed! Map was found not to be a blizzard map!")
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != true {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != true {
		return false
	}

	if gameDescription.MaxPlayers() != 6 {
		return false
	}

	return true
}

func checkRanked4v4(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 8) {
		return false
	}

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		log.Error("Integrity failed! Map was found not to be a blizzard map!")
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != true {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != true {
		return false
	}

	if gameDescription.MaxPlayers() != 8 {
		return false
	}

	return true
}

func checkRankedArchon(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 4) {
		return false
	}

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		log.Error("Integrity failed! Map was found not to be a blizzard map!")
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != true {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != true {
		return false
	}

	if gameDescription.MaxPlayers() != 4 {
		return false
	}

	return true
}

func checkCustom1v1(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 2) {
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != false {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != false {
		return false
	}

	if gameDescription.MaxPlayers() != 2 {
		return false
	}

	return true
}

func checkCustom2v2(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 4) {
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != false {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != false {
		return false
	}

	if gameDescription.MaxPlayers() != 4 {
		return false
	}

	return true
}

func checkCustom3v3(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 6) {
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != false {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != false {
		return false
	}

	if gameDescription.MaxPlayers() != 6 {
		return false
	}

	return true
}

func checkCustom4v4(replayData *rep.Rep) bool {

	if !checkNumberOfPlayers(replayData, 8) {
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != false {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != false {
		return false
	}

	if gameDescription.MaxPlayers() != 8 {
		return false
	}

	return true
}

func checkCustomFFA(replayData *rep.Rep) bool {

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions

	if gameOptions.Amm() != false {
		return false
	}

	if gameOptions.CompetitiveOrRanked() != false {
		return false
	}

	if gameDescription.MaxPlayers() != 8 {
		return false
	}

	return true
}
