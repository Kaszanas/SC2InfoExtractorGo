package dataproc

import (
	"encoding/json"

	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// TODO: Prepare anonymization using native golang structures
// Anonymization is needed in chat events and in Toon of the player.
// Players should receive persistent anonymized ID for every toon that was observed in the replay to be able to perform more advanced analysis.

// StringifyReplay allows the replayFile to be turned into a JSON data while cleaning the structure and anonymizing it.
func StringifyReplay(replayFile string) (bool, string) {

	log.Info("Entered stringifyReplay()")
	successFlag := true

	replayData, err := rep.NewFromFile(replayFile)
	if err != nil {
		log.WithFields(log.Fields{"file": replayFile, "error": err, "readError": true}).Error("Failed to read file.")
		return !successFlag, ""
	}
	defer replayData.Close()
	log.WithField("file", replayFile).Info("Read data from a replay.")

	cleanReplayData, redefError := redifineReplayStructure(replayData)
	if !redefError {
		log.WithField("file", replayFile).Error("Error in redefining replay structure.")
	}

	replayDataString, marshalErr := json.MarshalIndent(cleanReplayData, "", " ")
	if marshalErr != nil {
		log.WithField("file", replayFile).Error("Error while marshaling the string representation of cleanReplayData.")
	}

	// TODO: Return a summary in a custom Golang struct.
	return successFlag, string(replayDataString)
}
