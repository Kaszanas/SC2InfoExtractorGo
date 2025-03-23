package game_events

import log "github.com/sirupsen/logrus"

func CleanCameraSaveEvent(
	gameEventJSONMap map[string]any,
) {

	log.Debug("Entered cleanCameraUpdateEvent()")

	// TODO: recalculate camera coordinates (if needed)
	// https: //github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/camera/TargetPoint.java#L41
	// https://github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/camera/CameraUpdateEvent.java#L52

	// REVIEW: Camera TargetPoint has different recalculation than Cmd TargetPoint
	recalculateCameraTargetPoint(gameEventJSONMap)

	log.Debug("Finished cleanCameraUpdateEvent()")
}
