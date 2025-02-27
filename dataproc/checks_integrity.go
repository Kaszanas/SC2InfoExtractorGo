package dataproc

import (
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// Integrity:
// checkIntegrity verifies if the internal saved state of the replayData
// matches against structures with redundant information.
func checkIntegrity(replayData *rep.Rep) (bool, string) {

	log.Debug("Entered checkIntegrity()")
	maxPlayers := replayData.InitData.GameDescription.MaxPlayers()
	replayDetails := replayData.Details

	// Checking that the duration of the game is not equal to 0:
	if replayData.Header.Duration().Seconds() == 0 && replayData.Metadata.DurationSec() == 0 {
		log.WithFields(log.Fields{
			"headerDurationSeconds":   replayData.Header.Duration().Seconds(),
			"metadataDurationSeconds": replayData.Metadata.DurationSec(),
		}).Error("Integrity check failed! Detected the time of the game to be 0!")
		return false, "Both fields containing game time are empty!"
	}

	// Checking if the game version is not empty:
	if replayData.Metadata.GameVersion() == "" && replayData.Header.VersionString() == "" {
		log.WithFields(log.Fields{
			"metadataGameVersion": replayData.Metadata.GameVersion(),
			"headerGameVersion":   replayData.Header.VersionString(),
		}).Error("Integrity check failed! Detected game version to be empty!")
		return false, "Both fields containing game version are empty!"
	}

	// Technically there cannot be more than 15 human players!
	// Based on: https://s2editor-tutorials.readthedocs.io/en/master/01_Introduction/009_Player_Properties.html
	if maxPlayers > 16 || maxPlayers < 1 {
		log.WithField("maxPlayers", maxPlayers).
			Error("Integrity check failed! maxPlayers is not within the legal game engine range!")
		return false, "Player number doesn't fit within the maximum or minimum!"
	}

	// Map name of a replay is available in two places in the parsed data,
	// if they mismatch then first part of integrity check test fails:
	if replayData.Metadata.Title() != replayDetails.Title() {
		// Checking if both structures holding map name are empty:
		if replayData.Metadata.Title() == "" && replayDetails.Title() == "" {
			log.WithFields(log.Fields{
				"metadataTitle":      replayData.Metadata.Title(),
				"replayDetailsTitle": replayDetails.Title()}).
				Error("Integrity check failed! metadataTitle does not match replayDetailsTitle!")
			return false, "Both fields containing map name are empty!"
		}
	}

	// Checking if player list from replayDetails is of the same length as ToonPlayerDescMap:
	replayDetailsPlayerListLength := len(replayDetails.Players())
	toonPlayerDescMapLength := len(replayData.TrackerEvts.ToonPlayerDescMap)
	if replayDetailsPlayerListLength != toonPlayerDescMapLength {
		log.WithFields(log.Fields{
			"replayDetailsPlayerListLength": replayDetailsPlayerListLength,
			"toonPlayerDescMapLength":       toonPlayerDescMapLength}).
			Error("Integrity check failed! length of players mismatch!")
		return false, "Player lists length mismatch!"
	}

	gameDescIsBlizzardMap := replayData.InitData.GameDescription.IsBlizzardMap()
	detailsIsBlizzardMap := replayData.Details.IsBlizzardMap()

	// Checking if isBlizzardMap is the same in both of the available places:
	log.Info("Checking if the map included is marked as isBlizzardMap!")
	if gameDescIsBlizzardMap != detailsIsBlizzardMap {
		log.Error("Integrity failed! isBlizzardMap information is inconsistent within a processed file!")
		return false, "Two fields containing info if the map isBlizzardMap are different!"
	}

	log.Info("Integrity checks passed! Returning from checkIntegrity()")
	return true, ""
}
