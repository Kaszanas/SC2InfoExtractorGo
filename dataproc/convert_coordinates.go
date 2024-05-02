package dataproc

import (
	"encoding/json"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot"
	log "github.com/sirupsen/logrus"
)

// convertCoordinates accesses the data from GameEvents
// and recalculates the x,y,z coordinates of events
func convertCoordinates(replayData *replay_data.CleanedReplay) bool {
	log.Info("Entered convertCoordinates()")

	var newSliceOfEvents []s2prot.Struct

	for _, gameEvent := range replayData.GameEvents {

		targetCoordinates := gameEvent.String()
		var structInInterface map[string]interface{}
		err := json.Unmarshal([]byte(targetCoordinates), &structInInterface)
		if err != nil {
			log.Error("Failed to unmarshal the s2prot.Struct from GameEvents")
			return false
		}

		if val, ok := structInInterface["target"]; ok {
			// Check if target struct is not empty to avoid panics
			if val == nil {
				continue
			}
			assertedTarget := val.(map[string]interface{})

			// Check if the target contains x,y,z
			if val, ok := assertedTarget["x"]; ok {
				assertedTarget["x"] = val.(float64) / 8192.
			}
			if val, ok := assertedTarget["y"]; ok {
				assertedTarget["y"] = val.(float64) / 8192.
			}
			if val, ok := assertedTarget["z"]; ok {
				assertedTarget["z"] = val.(float64) / 8192.
			}
			structInInterface["target"] = assertedTarget
			// If not check if the target contains snapshotPoint with x,y,z inside
			if val, ok := assertedTarget["snapshotPoint"]; ok {
				// Check if the field is not nil in order to avoid panics
				if val == nil {
					continue
				}

				assertedSnapshotPoint := val.(map[string]interface{})
				// Converting coordinates
				if val, ok := assertedSnapshotPoint["x"]; ok {
					assertedSnapshotPoint["x"] = val.(float64) / 8192.
				}
				if val, ok := assertedSnapshotPoint["y"]; ok {
					assertedSnapshotPoint["y"] = val.(float64) / 8192.
				}
				if val, ok := assertedSnapshotPoint["z"]; ok {
					assertedSnapshotPoint["z"] = val.(float64) / 8192.
				}
				assertedTarget["snapshotPoint"] = assertedSnapshotPoint
			}

		}
		newSliceOfEvents = append(newSliceOfEvents, structInInterface)
	}

	replayData.GameEvents = newSliceOfEvents
	log.Info("Finished convertCoordinates()")

	return true
}
