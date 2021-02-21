package dataproc

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func Pipeline(replayFile string) (bool, string, data.ReplaySummary) {

	successFlag := true

	// Read replay:
	replayData, err := rep.NewFromFile(replayFile)
	if err != nil {
		log.WithFields(log.Fields{"file": replayFile, "error": err, "readError": true}).Error("Failed to read file.")
		return !successFlag, "", data.ReplaySummary{}
	}
	defer replayData.Close()
	log.WithField("file", replayFile).Info("Read data from a replay.")

	// TODO: Perform integrity checks

	// Clean replay structure:
	cleanOk, cleanReplayStructure := cleanReplay(replayData)
	if !cleanOk {
		log.WithField("file", replayFile).Error("Failed to perform cleaning.")
		return !successFlag, "", data.ReplaySummary{}
	}

	// Create replay summary:
	summarizeOk, summarizedReplay := summarizeReplay(&cleanReplayStructure)
	if !summarizeOk {

		return !successFlag, "", data.ReplaySummary{}
	}

	// TODO: Anonimize:

	// Create final replay string:
	stringifyOk, finalReplayString := stringifyReplay(&cleanReplayStructure)
	if !stringifyOk {
		log.WithField("file", replayFile).Error("Failed to stringify the replay.")
		return !successFlag, "", data.ReplaySummary{}
	}

	return successFlag, finalReplayString, summarizedReplay
}
