package game_events

import log "github.com/sirupsen/logrus"

// CleanCmdUpdateTargetUnitEvent cleans the CmdUpdateTargetUnit event.
// It recalculates the target field containing the snapshotPoint coordinates.
// The recalculated coordinates mutate the gameEventJSONMap map.
func CleanCmdUpdateTargetUnitEvent(gameEventJSONMap map[string]any) {
	// REVIEW: This event is not cleaned, should it be cleaned?

	if target, ok := gameEventJSONMap["target"]; ok {
		if target == nil {
			log.Debug("Detected nil game event target")
		} else {
			castedTarget := target.(map[string]any)

			if snapshotPoint, ok := castedTarget["snapshotPoint"]; ok {
				if snapshotPoint == nil {
					log.Debug("Detected nil game event snapshotPoint")
				} else {
					castedSnapshotPoint := snapshotPoint.(map[string]any)

					// REVIEW: Values seem to be extremely small after recalculation:
					if val, ok := castedSnapshotPoint["x"]; ok {
						castedSnapshotPoint["x"] = val.(float64) / 8192.
					}
					if val, ok := castedSnapshotPoint["y"]; ok {
						castedSnapshotPoint["y"] = val.(float64) / 8192.
					}
					if val, ok := castedSnapshotPoint["z"]; ok {
						castedSnapshotPoint["z"] = val.(float64) / 8192.
					}
					castedTarget["snapshotPoint"] = castedSnapshotPoint
				}
			}
		}
	}
}
