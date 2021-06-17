package dataproc

import (
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func checkIntegrity(replayData *rep.Rep) bool {

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if replayData.InitData.GameDescription.Struct["isBlizzardMap"].(bool) == replayData.Details.IsBlizzardMap() {
		log.Error("Integrity failed! Map was found not to be a blizzard map!")
		return false
	}

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

	return true
}

func checkCompetitiveRanked1v1(replayData *rep.Rep) bool {
	// TODO: Check if the replay is competitive 1v1
	// Within the dataset that is being prepared that should be the case but otherwise this software should be universal.

	return true
}
