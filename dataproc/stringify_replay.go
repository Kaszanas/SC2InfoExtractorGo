package dataproc

import (
	"encoding/json"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	log "github.com/sirupsen/logrus"
)

// stringifyReplay performs marshaling of all of CleanedReplay information into a string.
func stringifyReplay(replayData *replay_data.CleanedReplay) (bool, string) {

	log.Debug("Entered stringifyReplay()")

	replayDataString, marshalErr := json.MarshalIndent(replayData, "", "  ")
	if marshalErr != nil {
		log.Error("Error while marshaling the string representation of cleanReplayData.")
		return false, ""
	}

	log.Debug("Finished stringifyReplay()")
	return true, string(replayDataString)
}
