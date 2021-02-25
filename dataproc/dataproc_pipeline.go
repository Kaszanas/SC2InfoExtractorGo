package dataproc

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// Pipeline is performing the whole data processing pipeline for a replay file. Reads the replay, cleans the replay structure, creates replay summary, anonymizes, and creates a JSON replay output.
func Pipeline(replayFile string) (bool, string, data.ReplaySummary) {

	// Read replay:
	replayData, err := rep.NewFromFile(replayFile)
	if err != nil {
		log.WithFields(log.Fields{"file": replayFile, "error": err, "readError": true}).Error("Failed to read file.")
		return false, "", data.ReplaySummary{}
	}
	log.WithField("file", replayFile).Info("Read data from a replay.")

	// TODO: Perform integrity checks

	// Clean replay structure:
	cleanOk, cleanReplayStructure := cleanReplay(replayData)
	if !cleanOk {
		log.WithField("file", replayFile).Error("Failed to perform cleaning.")
		return false, "", data.ReplaySummary{}
	}

	// Create replay summary:
	summarizeOk, summarizedReplay := summarizeReplay(&cleanReplayStructure)
	if !summarizeOk {
		log.WithField("file", replayFile).Error("Failed to create replay summary.")
		return false, "", data.ReplaySummary{}
	}

	// Anonimize replay:
	if !anonymizeReplay(&cleanReplayStructure) {
		log.WithField("file", replayFile).Error("Failed to anonymize replay.")
		return false, "", data.ReplaySummary{}
	}

	// Create final replay string:
	stringifyOk, finalReplayString := stringifyReplay(&cleanReplayStructure)
	if !stringifyOk {
		log.WithField("file", replayFile).Error("Failed to stringify the replay.")
		return false, "", data.ReplaySummary{}
	}

	replayData.Close()

	return true, finalReplayString, summarizedReplay
}
