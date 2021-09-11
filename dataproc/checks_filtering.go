package dataproc

import (
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// Filtering
// filterGameModes performs the check against a binary 0b1111111 getGameModeFlag to verify if the currently processed replay game mode is correct.
func filterGameModes(replayData *rep.Rep, getGameModeFlag int) bool {
	log.Info("Entered checkGameMode()")

	for _, value := range gameModeList {
		// If we want to include games with game mode `value`, and the game matches the requirements
		// of the game mode, then it matches the filter => return true.
		if getGameModeFlag&value != 0 && checkGameParameters(replayData, gameModeFiltersMapping[value]) {
			return true
		}
	}

	log.Info("Finished checkGameMode()")

	// The game did not match any active filters, return false.
	return false
}

// checkGameParameters takes in a VerifyGameInfo struct that containts information about specific game mode filtering based on available data in the replay file:
func checkGameParameters(replayData *rep.Rep, gameInfoFilter VerifyGameInfo) bool {

	log.Info("Entered checkGameParameters()")

	// Verifying if the number of players matches:
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

	playerList := replayData.Metadata.Players()
	numberOfPlayers := len(playerList)

	log.WithFields(log.Fields{
		"len(playerList)": numberOfPlayers,
		"requiredNumber":  requiredNumber}).Debug("checkNumberOfPlayers()")

	if numberOfPlayers != requiredNumber {
		return false
	}

	return true
}
