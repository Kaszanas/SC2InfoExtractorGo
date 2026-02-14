package dataproc

import (
	"bytes"
	"encoding/json"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	log "github.com/sirupsen/logrus"
)

// stringifyReplay performs marshaling of all of CleanedReplay information into a string.
func stringifyReplay(replayData *replay_data.CleanedReplay) (bool, string) {

	log.Debug("Entered stringifyReplay()")


	replayDataStringBytes, marshalErr := json.Marshal(replayData)
	if marshalErr != nil {
		log.Error("Error while marshaling the string representation of cleanReplayData.")
		return false, ""
	}

	compactedOutput := new(bytes.Buffer)
	compactErr := json.Compact(compactedOutput, replayDataStringBytes)
	if compactErr != nil {
		log.Error("Error while compacting the string representation of cleanReplayData.")
		return false, ""
	}

	compactedString := compactedOutput.String()

	log.Debug("Finished stringifyReplay()")
	return true, compactedString
}
