package dataproc

import (
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

type GameMode int

const (
	AllGameModes GameMode = iota + 1
	Ranked1v1
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

func (w GameMode) String() string {
	return [...]string{"AllGameModes", "Ranked1v1", "Ranked2v2", "Ranked3v3", "Ranked4v4", "RankedArchon", "Custom1v1", "Custom2v2", "Custom3v3", "Custom4v4", "CustomFFA"}[w-1]
}

func (w GameMode) EnumIndex() int {
	return int(w)
}

func checkIntegrity(replayData *rep.Rep, checkIntegrityBool bool, checkGameMode int) bool {

	if checkGameMode == AllGameModes.EnumIndex() {
		if checkIntegrityBool {
			basicIntegrityOk := checkBasicIntegrity(replayData)
			if !basicIntegrityOk {
				return false
			}
		}
		return true
	}

	if checkGameMode == Ranked1v1.EnumIndex() {
		is1v1RankedGameMode := checkRanked1v1(replayData)
		if !is1v1RankedGameMode {
			return false
		}

		if checkIntegrityBool {
			basicIntegrityOk := checkBasicIntegrity(replayData)
			if !basicIntegrityOk {
				return false
			}
		}
	}

	// TODO: check which game mode is currently being processed and if the integrity is to be upheld.
	if checkIntegrityBool {
		basicIntegrityOk := checkBasicIntegrity(replayData)
		if !basicIntegrityOk {
			return false
		}
	}

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		log.Error("Integrity failed! Map was found not to be a blizzard map!")
		return false
	}

	return true
}

func checkBasicIntegrity(replayData *rep.Rep) bool {
	// Check gameEvents "userOptions" "buildNum" and "baseBuildNum" against "header" information:
	playerList := replayData.Metadata.Players()

	if len(playerList) < 2 {
		log.Error("Integrity check failed number of players is less than 2!")
		return false
	}

	// MMR should be below 8000 for all of the replays:
	for _, playerStats := range playerList {

		// TODO: Encode maximum MMR difference between players that is possible in the game:
		// Around 1200 MMR
		if playerStats.MMR() > 8000 {
			log.Error("Integrity check failed! One of the players MMR is higher than 8000!")
			return false
		}

		if playerStats.APM() == 0 {
			log.Error("Integrity check failed! One of the players APM is equal to 0!")
			return false
		}
	}

}

func checkGameMode() {

}

func checkRanked1v1(replayData *rep.Rep) bool {

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
