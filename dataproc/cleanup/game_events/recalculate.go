package game_events

import log "github.com/sirupsen/logrus"

// recalculateCmdTargetUnitSnapshotPoint recalculates the snapshotPoint coordinates
// for the targetUnit and returns the mutated gameEventData.
func recalculateCmdTargetUnitSnapshotPoint(
	gameEventJSONMap map[string]any,
) {

	log.Debug("Entered recalculateCmdSnapshotPoint()")

	if targetUnit, ok := gameEventJSONMap["TargetUnit"]; ok {
		if targetUnit == nil {
			log.Debug("Detected nil targetUnit")
		} else {

			castedTargetUnit := targetUnit.(map[string]any)

			if cmdSnapshotPoint, ok := castedTargetUnit["snapshotPoint"]; ok {
				if cmdSnapshotPoint == nil {
					log.Debug("Detected nil snapshotPoint")
				} else {
					castedCmdSnapshotPoint := cmdSnapshotPoint.(map[string]any)

					nBits := 13
					divisor := 1 << nBits

					if val, ok := castedCmdSnapshotPoint["x"]; ok {
						if val != nil {
							castedCmdSnapshotPoint["x"] = val.(float64) / float64(divisor)
						}
					}
					if val, ok := castedCmdSnapshotPoint["y"]; ok {
						if val != nil {
							castedCmdSnapshotPoint["y"] = val.(float64) / float64(divisor)
						}
					}
					if val, ok := castedCmdSnapshotPoint["z"]; ok {
						if val != nil {
							castedCmdSnapshotPoint["z"] = val.(float64) / float64(divisor)
						}
					}
					gameEventJSONMap["snapshotPoint"] = castedCmdSnapshotPoint
				}
			}
		}
	}

	log.Debug("Finished recalculateCmdSnapshotPoint()")
}

// recalculateCmdTargetPoint recalculates the targetPoint
// coordinates and returns the mutated gameEventData.
func recalculateCmdTargetPoint(
	gameEventJSONMap map[string]any,
) {

	log.Debug("Entered recalculateCmdTargetPoint()")

	// REVIEW: sc2reader is not recalculating this coordinate,
	// at the same time scelight provides a recalculated float value:
	// - https://github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/cmd/TargetPoint.java#L41
	// - https://github.com/ggtracker/sc2reader/blob/ba8b52ec0875df5cd869af09dccdb4d9f84ae921/sc2reader/events/game.py#L267-L292

	if targetPoint, ok := gameEventJSONMap["TargetPoint"]; ok {
		if targetPoint == nil {
			log.Debug("Detected nil targetPoint")
		} else {
			nBits := 13
			divisor := 1 << nBits

			castedTargetPoint := targetPoint.(map[string]any)

			// REVIEW: Lots of code repetition for these:
			if val, ok := castedTargetPoint["x"]; ok {
				if val == nil {
					log.Debug("Detected nil x value")
				} else {
					castedTargetPoint["x"] = val.(float64) / float64(divisor)
				}
			}
			if val, ok := castedTargetPoint["y"]; ok {
				if val == nil {
					log.Debug("Detected nil y value")
				} else {
					castedTargetPoint["y"] = val.(float64) / float64(divisor)
				}
			}
			if val, ok := castedTargetPoint["z"]; ok {
				if val == nil {
					log.Debug("Detected nil z value")
				} else {
					castedTargetPoint["z"] = val.(float64) / float64(divisor)
				}
			}
			gameEventJSONMap["TargetPoint"] = castedTargetPoint
		}
	}

}

// func recalculateGameEventTarget(gameEventJSONMap map[string]any) {
// 	log.Debug("Entered recalculateGameEventTargetPoint()")

// 	// REVIEW: Values seem to be extremely small after recalculation:
// 	if target, ok := gameEventJSONMap["target"]; ok && target != nil {

// 		castedTarget := target.(map[string]any)

// 		if val, ok := castedTarget["x"]; ok {
// 			castedTarget["x"] = val.(float64) / 8192.
// 		}
// 		if val, ok := castedTarget["y"]; ok {
// 			castedTarget["y"] = val.(float64) / 8192.
// 		}
// 		if val, ok := castedTarget["z"]; ok {
// 			castedTarget["z"] = val.(float64) / 8192.
// 		}
// 		gameEventJSONMap["target"] = castedTarget
// 	}
// 	log.Debug("Finished recalculateGameEventTargetPoint()")
// }

func recalculateCameraTargetPoint(
	gameEventJSONMap map[string]any,
) {

	log.Debug("Entered recalculateCameraTargetPoint()")

	if target, ok := gameEventJSONMap["target"]; ok {
		if target == nil {
			log.Debug("Detected nil target")
		} else {

			castedTargetPoint := target.(map[string]any)
			if castedTargetPoint == nil {
				log.Debug("Detected nil target point")
			}

			if val, ok := castedTargetPoint["x"]; ok {
				castedTargetPoint["x"] = val.(float64) / 256.0
			}
			if val, ok := castedTargetPoint["y"]; ok {
				castedTargetPoint["y"] = val.(float64) / 256.0
			}

			gameEventJSONMap["target"] = castedTargetPoint

		}
	}

	// return distance == null ? null : distance / 256.0f;
	if distance, ok := gameEventJSONMap["distance"]; ok {
		if distance == nil {
			log.Debug("Detected nil distance")
		} else {
			castedDistance := gameEventJSONMap["distance"].(float64) / 256.0
			gameEventJSONMap["distance"] = castedDistance
		}

	}

	log.Debug("Finished recalculateCameraTargetPoint()")
}

func recalculatePitchYaw(
	gameEventJSONMap map[string]any,
) {
	log.Debug("Entered recalculatePitchYaw()")

	// Recalculate pitch to degrees
	if pitch, ok := gameEventJSONMap["pitch"]; ok {
		if pitch == nil {
			log.Debug("Detected nil pitch")
		} else {
			castedPitch := int64(pitch.(float64))
			// return pitch == null ? null : ( 45 * ( ( ( ( ( ( pitch << 5 ) - 0x2000 ) << 17 ) - 1 ) >> 17 ) + 1 ) ) / 4096.0f;
			recalculatedPitch := float64((45 * ((((((castedPitch << 5) - 0x2000) << 17) - 1) >> 17) + 1))) / 4096.0

			gameEventJSONMap["pitch"] = recalculatedPitch
		}
	}

	// Recalculate yaw to degrees
	if yaw, ok := gameEventJSONMap["yaw"]; ok {
		if yaw == nil {
			log.Debug("Detected nil yaw")
		} else {
			castedYaw := int64(yaw.(float64))
			// return yaw == null ? null : ( 45 * ( ( ( ( ( ( yaw << 5 ) - 0x2000 ) << 17 ) - 1 ) >> 17 ) + 1 ) ) / 4096.0f;
			recalculatedYaw := float64((45 * ((((((castedYaw << 5) - 0x2000) << 17) - 1) >> 17) + 1))) / 4096.0
			gameEventJSONMap["yaw"] = recalculatedYaw
		}
	}

	log.Debug("Finished recalculatePitchYaw()")
}
