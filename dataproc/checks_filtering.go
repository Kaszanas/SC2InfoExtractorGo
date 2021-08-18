package dataproc

import (
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// Filtering
// checkGameMode performs the check against a binary 0b1111111 getGameModeFlag to verify if the currently processed replay game mode is correct.
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
		log.Info("Filtering game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions
	gameOptionsAmm := gameOptions.Amm()

	if gameOptionsAmm != gameInfoFilter.isAutoMatchMaking {
		log.WithFields(log.Fields{
			"gameOptionsAmm":    gameOptionsAmm,
			"isAutoMatchMaking": gameInfoFilter.isAutoMatchMaking}).Info("Filtering game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	competitiveOrRanked := gameOptions.CompetitiveOrRanked()
	if competitiveOrRanked != gameInfoFilter.isCompetitiveOrRanked {
		log.WithFields(log.Fields{
			"competitiveOrRanked":   competitiveOrRanked,
			"isCompetitiveOrRanked": gameInfoFilter.isCompetitiveOrRanked}).Info("Filtering game parameters mismatch! returning from checkGameParameters()")
		return false
	}

	maxPlayers := gameDescription.MaxPlayers()
	if maxPlayers != int64(gameInfoFilter.maxPlayers) {
		log.WithFields(log.Fields{
			"maxPlayers":                maxPlayers,
			"gameInfoFilter.maxPlayers": gameInfoFilter.maxPlayers}).Info("Filtering game parameters mismatch! returning from checkGameParameters()")
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
		log.WithFields(log.Fields{
			"len(playerList)": numberOfPlayers,
			"requiredNumber":  requiredNumber}).Info("Different number of players than required number")
		return false
	}

	log.Info("Finished checkNumberOfPlayers(), returning")
	return true
}
