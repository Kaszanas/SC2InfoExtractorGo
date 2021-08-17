package dataproc

import (
	"encoding/json"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	log "github.com/sirupsen/logrus"
)

// stringifyReplay performs marshaling of all of CleanedReplay information into a string.
func stringifyReplay(replayData *data.CleanedReplay) (bool, string) {

	log.Info("Entered stringifyReplay()")
	successFlag := true

	replayDataString, marshalErr := json.MarshalIndent(replayData, "", "  ")
	if marshalErr != nil {
		log.Error("Error while marshaling the string representation of cleanReplayData.")
		return !successFlag, ""
	}

	log.Info("Finished stringifyReplay()")
	return successFlag, string(replayDataString)
}
