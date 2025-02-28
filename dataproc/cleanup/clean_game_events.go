package cleanup

import (
	"encoding/json"

	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

const (
	Alternate              = 1 << iota // 0x01
	Queued                             // 0x02
	Preempt                            // 0x04
	SmartClick                         // 0x08
	SmartRally                         // 0x10
	Subgroup                           // 0x20
	SetAutoCastOff                     // 0x40
	SetAutoCastOn                      // 0x80
	User                               // 0x100
	DataA                              // 0x200
	DataB                              // 0x400
	AI                                 // 0x800
	AIIgnoreOnFinish                   // 0x1000
	Order                              // 0x2000
	Script                             // 0x4000
	HomogenousInterruption             // 0x8000
	Minimap                            // 0x10000
	Repeat                             // 0x20000
	DispatchToOtherUnit                // 0x40000
	TargetSelf                         // 0x80000
)

// cleanGameEvents copies the game events,
// has the capability of removing unecessary fields.
func CleanGameEvents(replayData *rep.Rep) []map[string]interface{} {
	log.Debug("Entered CleanGameEvents()")

	// Constructing a clean GameEvents without unescessary fields:
	// NOTE: cleanGameEvents are of type map[string]interface{} because to effectively
	// add, recalculate or remove field from the game events it is easier to work with
	// maps than with s2prot.Structs, maps are related closer to the final JSON format.
	var cleanGameEvents []map[string]interface{}
	for _, gameEvent := range replayData.GameEvts {

		gameEventStruct, err := cleanGameEvent(gameEvent)
		if err != nil {
			log.Error("Failed to clean GameEvent, event will stay in the raw format")
		}

		cleanGameEvents = append(cleanGameEvents, gameEventStruct)
	}

	log.Debug("Finished cleaning GameEvents")
	return cleanGameEvents
}

// cleanGameEvent is responsible for unmarshalling the string representation of a
// s2prot.Struct from the game events, mutating the struct by adding/removing/recalculating
// fields and returning the mutated struct.
func cleanGameEvent(gameEvent s2prot.Event) (map[string]interface{}, error) {
	log.Debug("Entered cleanGameEvent()")

	gameEventBytes := gameEvent.Struct.String()
	var gameEventJSONMap map[string]interface{}
	err := json.Unmarshal([]byte(gameEventBytes), &gameEventJSONMap)
	if err != nil {
		log.Error("Failed to unmarshal the s2prot.Struct from GameEvents")
		return nil, err
	}

	// REVIEW: Verify if there are any more game event types that need to be cleaned:
	// recalculateGameEventTarget(gameEventJSONMap)

	if gameEvent.Name == "Cmd" {
		cleanCmdEvent(gameEvent, gameEventJSONMap)
	}

	if gameEvent.Name == "CameraUpdate" {
		cleanCameraUpdateEvent(gameEventJSONMap)
	}

	log.Debug("Finished cleanGameEvent()")
	return gameEventJSONMap, nil

}

// func recalculateGameEventTarget(gameEventJSONMap map[string]interface{}) {
// 	log.Debug("Entered recalculateGameEventTargetPoint()")

// 	// REVIEW: Values seem to be extremely small after recalculation:
// 	if target, ok := gameEventJSONMap["target"]; ok && target != nil {

// 		castedTarget := target.(map[string]interface{})

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

func cleanCmdEvent(
	gameEvent s2prot.Event,
	gameEventJSONMap map[string]interface{},
) {

	log.Debug("Entered cleanCmdEvent()")

	if gameEventAbility, ok := gameEventJSONMap["abil"]; ok && gameEventAbility != nil {
		castedGameEventAbility := gameEventAbility.(map[string]interface{})
		// Acquiring ability data:
		getAbilityData(castedGameEventAbility)
		gameEventJSONMap["abil"] = castedGameEventAbility
	}

	// Recalculating snapshotPoint for a TargetUnit:
	if gameEventData, ok := gameEventJSONMap["data"]; ok && gameEventData != nil {

		castedGameEventData := gameEventData.(map[string]interface{})

		// Recalculating SnapshotPoint for a TargetUnit:
		recalculateCmdSnapshotPoint(castedGameEventData)

		// Recalculating TargetPoint:
		recalculateCmdTargetPoint(castedGameEventData)

		gameEventJSONMap["data"] = castedGameEventData
	}

	// Getting command flags:
	getCmdFlags(gameEvent, gameEventJSONMap)

	log.Debug("Finished cleanCmdEvent()")
}

// recalculateCmdTargetPoint recalculates the targetPoint
// coordinates and returns the mutated gameEventData.
func recalculateCmdTargetPoint(
	gameEventJSONMap map[string]interface{},
) {

	log.Debug("Entered recalculateCmdTargetPoint()")

	// REVIEW: sc2reader is not recalculating this coordinate,
	// at the same time scelight provides a recalculated float value:
	// - https://github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/cmd/TargetPoint.java#L41
	// - https://github.com/ggtracker/sc2reader/blob/ba8b52ec0875df5cd869af09dccdb4d9f84ae921/sc2reader/events/game.py#L267-L292

	if targetPoint, ok := gameEventJSONMap["TargetPoint"]; ok {
		nBits := 13
		divisor := 1 << nBits

		castedTargetPoint := targetPoint.(map[string]interface{})

		// REVIEW: Lots of code repetition for these:
		if val, ok := castedTargetPoint["x"]; ok {
			if val != nil {
				castedTargetPoint["x"] = val.(float64) / float64(divisor)
			}
		}
		if val, ok := castedTargetPoint["y"]; ok {
			if val != nil {
				castedTargetPoint["y"] = val.(float64) / float64(divisor)
			}
		}
		if val, ok := castedTargetPoint["z"]; ok {
			if val != nil {
				castedTargetPoint["z"] = val.(float64) / float64(divisor)
			}
		}
		// targetPointZ := float64(targetPoint["z"].(int64))

		gameEventJSONMap["TargetPoint"] = castedTargetPoint
	}

}

// recalculateCmdSnapshotPoint recalculates the snapshotPoint coordinates
// for the targetUnit and returns the mutated gameEventData.
func recalculateCmdSnapshotPoint(
	gameEventJSONMap map[string]interface{},
) {

	log.Debug("Entered recalculateCmdSnapshotPoint()")

	if targetUnit, ok := gameEventJSONMap["TargetUnit"]; ok && targetUnit != nil {
		castedTargetUnit := targetUnit.(map[string]interface{})

		if cmdSnapshotPoint, ok := castedTargetUnit["snapshotPoint"]; ok && cmdSnapshotPoint != nil {

			castedCmdSnapshotPoint := cmdSnapshotPoint.(map[string]interface{})

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

	log.Debug("Finished recalculateCmdSnapshotPoint()")
}

// getCmdFlags translates the integer bitmask to a string array,
// and returns the mutated gameEventStruct.
func getCmdFlags(
	gameEvent s2prot.Event,
	gameEventJSONMap map[string]interface{},
) {

	log.Debug("Entered getCmdFlags()")

	gameEventFlagBitmask := gameEvent.Int("cmdFlags")
	commandFlags := getCommandType(gameEventFlagBitmask)

	log.WithField("CommandType", commandFlags).Debug("Got command type")
	log.Debug("GameEventInt: ", gameEventFlagBitmask)

	gameEventJSONMap["cmdFlags"] = commandFlags

	log.Debug("Finished getCmdFlags()")
}

func getCommandType(cmdFlags int64) []string {

	log.Debug("Entered getCommandType()")

	var flags []string
	// Scelight is ommiting User flag, all commands must have a user:
	if cmdFlags&User != 0 {
		flags = append(flags, "User")
	}

	if cmdFlags&Alternate != 0 {
		flags = append(flags, "Alternate")
	}
	if cmdFlags&Queued != 0 {
		flags = append(flags, "Queued")
	}
	if cmdFlags&Preempt != 0 {
		flags = append(flags, "Preempt")
	}
	if cmdFlags&SmartClick != 0 {
		flags = append(flags, "SmartClick")
	}
	if cmdFlags&SmartRally != 0 {
		flags = append(flags, "SmartRally")
	}
	if cmdFlags&Subgroup != 0 {
		flags = append(flags, "Subgroup")
	}
	if cmdFlags&SetAutoCastOff != 0 {
		flags = append(flags, "SetAutoCastOff")
	}
	if cmdFlags&SetAutoCastOn != 0 {
		flags = append(flags, "SetAutoCastOn")
	}

	if cmdFlags&DataA != 0 {
		flags = append(flags, "DataA")
	}
	if cmdFlags&DataB != 0 {
		flags = append(flags, "DataB")
	}
	if cmdFlags&AI != 0 {
		flags = append(flags, "AI")
	}
	if cmdFlags&AIIgnoreOnFinish != 0 {
		flags = append(flags, "AIIgnoreOnFinish")
	}
	if cmdFlags&Order != 0 {
		flags = append(flags, "Order")
	}
	if cmdFlags&Script != 0 {
		flags = append(flags, "Script")
	}
	if cmdFlags&HomogenousInterruption != 0 {
		flags = append(flags, "HomogenousInterruption")
	}
	if cmdFlags&Minimap != 0 {
		flags = append(flags, "Minimap")
	}
	if cmdFlags&Repeat != 0 {
		flags = append(flags, "Repeat")
	}
	if cmdFlags&DispatchToOtherUnit != 0 {
		flags = append(flags, "DispatchToOtherUnit")
	}
	if cmdFlags&TargetSelf != 0 {
		flags = append(flags, "TargetSelf")
	}

	log.Debug("Finished getCommandType()")

	return flags
}

// getAbilityData acquires the ability data from the gameEventAbility struct.
// Mutates the gameEventAbility struct and returns it containing
// a human readable ability name and command name.
func getAbilityData(gameEventJSONMap map[string]interface{}) {

	log.Debug("Entered getAbilityData()")

	// TODO: load this from balance data where possible
	// REVIEW: What to do if no balance data is available?
	// should this break the processing?
	// TODO: I have found the proper string ability name and the string ability command.
	// Either replace the values with strings,
	// maybe a better option is to add more fields such as abilityName and abilityCommandName
	getAbilityName(gameEventJSONMap)
	getAbilityCommandName(gameEventJSONMap)

	log.Debug("Finished getAbilityData()")
}

// getAbilityName acquires the ability name, mutates the gameEventAbility and returns it.
func getAbilityName(
	gameEventJSONMap map[string]interface{},
) map[string]interface{} {

	log.Debug("Entered getAbilityName()")

	// TODO: This requires lookup to the game version balance data

	// 	gameEventAbility := gameEventAbility.(s2prot.Struct)
	// 	abilityLink := gameEventAbility["abilLink"].(int64)

	log.Debug("Finished getAbilityName()")

	return gameEventJSONMap
}

// getAbilityCommandName acquires the ability specific command name,
// mutates the gameEventAbility and returns it.
func getAbilityCommandName(
	gameEventJSONMap map[string]interface{},
) map[string]interface{} {

	log.Debug("Entered getAbilityCommandName()")

	// TODO: This requires lookup to the game version balance data

	// 	abilityCmdIndex := gameEventAbility["abilCmdIndex"].(int64)
	// 	abilCmdData := gameEventAbility["abilCmdData"]
	// 	if abilCmdData != nil {
	// 		log.Info("abilCmdData is not nil")
	// 	}
	// 	log.WithField("abilityLink", abilityLink).Debug("Got ability link")
	// 	log.WithField("abilityCmdIndex", abilityCmdIndex).Debug("Got ability cmdIndex")
	// 	log.WithField("abilCmdData", abilCmdData).Debug("Got ability cmdData")

	log.Debug("Finished getAbilityCommandName()")

	return gameEventJSONMap
}

func cleanCameraUpdateEvent(
	gameEventJSONMap map[string]interface{},
) {

	log.Debug("Entered cleanCameraUpdateEvent()")

	// TODO: recalculate camera coordinates (if needed)
	// https: //github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/camera/TargetPoint.java#L41
	// https://github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/camera/CameraUpdateEvent.java#L52

	// REVIEW: Camera TargetPoint has different recalculation than Cmd TargetPoint
	recalculatePitchYaw(gameEventJSONMap)
	recalculateCameraTargetPoint(gameEventJSONMap)

	log.Debug("Finished cleanCameraUpdateEvent()")
}

func recalculatePitchYaw(
	gameEventJSONMap map[string]interface{},
) {
	log.Debug("Entered recalculatePitchYaw()")

	// Recalculate pitch to degrees
	if pitch, ok := gameEventJSONMap["pitch"]; ok {
		if pitch != nil {
			castedPitch := int64(pitch.(float64))
			// return pitch == null ? null : ( 45 * ( ( ( ( ( ( pitch << 5 ) - 0x2000 ) << 17 ) - 1 ) >> 17 ) + 1 ) ) / 4096.0f;
			recalculatedPitch := float64((45 * ((((((castedPitch << 5) - 0x2000) << 17) - 1) >> 17) + 1))) / 4096.0

			gameEventJSONMap["pitch"] = recalculatedPitch
		}
	}

	// Recalculate yaw to degrees
	if yaw, ok := gameEventJSONMap["yaw"]; ok {
		if yaw != nil {
			castedYaw := int64(yaw.(float64))
			// return yaw == null ? null : ( 45 * ( ( ( ( ( ( yaw << 5 ) - 0x2000 ) << 17 ) - 1 ) >> 17 ) + 1 ) ) / 4096.0f;
			recalculatedYaw := float64((45 * ((((((castedYaw << 5) - 0x2000) << 17) - 1) >> 17) + 1))) / 4096.0
			gameEventJSONMap["yaw"] = recalculatedYaw
		}
	}

	log.Debug("Finished recalculatePitchYaw()")
}

func recalculateCameraTargetPoint(
	gameEventJSONMap map[string]interface{},
) {

	log.Debug("Entered recalculateCameraTargetPoint()")

	if targetPoint, ok := gameEventJSONMap["target"]; ok && targetPoint != nil {

		castedTargetPoint := targetPoint.(map[string]interface{})

		if val, ok := castedTargetPoint["x"]; ok {
			castedTargetPoint["x"] = val.(float64) / 256.0
		}
		if val, ok := castedTargetPoint["y"]; ok {
			castedTargetPoint["y"] = val.(float64) / 256.0
		}

		gameEventJSONMap["target"] = castedTargetPoint

	}

	// return distance == null ? null : distance / 256.0f;

	if distance, ok := gameEventJSONMap["distance"]; ok && distance != nil {

		castedDistance := gameEventJSONMap["distance"].(float64) / 256.0
		gameEventJSONMap["distance"] = castedDistance
	}

	log.Debug("Finished recalculateCameraTargetPoint()")
}
