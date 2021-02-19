package dataproc

import (
	"encoding/json"

	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

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

	structuredReplayData, redefOk := redifineReplayStructure(replayData)
	if !redefOk {
		log.WithField("file", replayFile).Error("Error in redefining replay structure.")
		return !successFlag, ""
	}

	// TODO: This needs to be controlled from outside of stringify_replay in case other users don't want to receive clean data.
	cleaningOk := cleanReplayStructure(&structuredReplayData)
	if !cleaningOk {
		log.WithField("file", replayFile).Error("Error in cleaning the replay structure.")
		return !successFlag, ""
	}

	replayDataString, marshalErr := json.MarshalIndent(structuredReplayData, "", "  ")
	if marshalErr != nil {
		log.WithField("file", replayFile).Error("Error while marshaling the string representation of cleanReplayData.")
		return !successFlag, ""
	}

	// TODO: Return a summary in a custom Golang struct.
	return successFlag, string(replayDataString)
}
