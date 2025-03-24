package game_events

import log "github.com/sirupsen/logrus"

// CleanCmdUpdateTargetPointEvent cleans the CmdUpdateTargetPoint game event.
// This includes the recalculation of the target point coordinates.
// Reference: https://github.com/icza/scelight/blob/master/src-app/hu/scelight/sc2/rep/model/gameevents/cmd/TargetPoint.java#L29
// Command TargetPoints seem to have to be recalculated by dividing the x, y, z coordinates by 8192.
func CleanCmdUpdateTargetPointEvent(gameEventJSONMap map[string]any) {

	if target, ok := gameEventJSONMap["target"]; ok {
		if target == nil {
			log.Debug("Detected nil game event target")
		} else {
			castedTarget := target.(map[string]any)

			if x, ok := castedTarget["x"]; ok {
				castedTarget["x"] = x.(float64) / 8192.
			}
			if y, ok := castedTarget["y"]; ok {
				castedTarget["y"] = y.(float64) / 8192.
			}
			if z, ok := castedTarget["z"]; ok {
				castedTarget["z"] = z.(float64) / 8192.
			}

			gameEventJSONMap["target"] = castedTarget

		}

	}
}
