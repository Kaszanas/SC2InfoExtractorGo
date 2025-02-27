package cleanup

import (
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

func getCommandType(cmdFlags int64) []string {

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

	return flags
}

// cleanGameEvents copies the game events,
// has the capability of removing unecessary fields.
func CleanGameEvents(replayData *rep.Rep) []s2prot.Struct {
	// Constructing a clean GameEvents without unescessary fields:
	var gameEventsStructs []s2prot.Struct
	for _, gameEvent := range replayData.GameEvts {

		gameEventStruct := gameEvent.Struct

		eventTypeName := gameEventStruct["evtTypeName"].(string)
		if eventTypeName == "Cmd" {
			gameEventStruct = cleanCmdEvent(gameEvent, gameEventStruct)
		}

		if eventTypeName == "CameraUpdate" {
			gameEventStruct = cleanCameraUpdateEvent(gameEventStruct)
		}

		gameEventsStructs = append(gameEventsStructs, gameEventStruct)
	}
	log.Info("Defined cleanGameEvents struct")
	return gameEventsStructs
}

func cleanCmdEvent(
	gameEvent s2prot.Event,
	gameEventStruct s2prot.Struct,
) s2prot.Struct {

	gameEventAbility := gameEventStruct["abil"]
	if gameEventAbility != nil {
		gameEventAbility := gameEventStruct["abil"].(s2prot.Struct)
		abilityData := getAbilityData(gameEventAbility)
		gameEventAbility["abilityData"] = abilityData
		gameEventStruct["abil"] = gameEventAbility
	}

	// Recalculating snapshotPoint for a TargetUnit:
	gameEventData := gameEventStruct["data"]
	if gameEventData != nil {
		gameEventData := gameEventStruct["data"].(s2prot.Struct)
		// Recalculating SnapshotPoint for a TargetUnit:
		gameEventData = recalculateCmdSnapshotPoint(gameEventData)

		// Recalculating TargetPoint:
		gameEventData = recalculateCmdTargetPoint(gameEventData)
		gameEventStruct["data"] = gameEventData
	}

	// Getting command flags:
	gameEventStruct = getCmdFlags(gameEvent, gameEventStruct)

	return gameEventStruct
}

// recalculateCmdTargetPoint recalculates the targetPoint
// coordinates and returns the mutated gameEventData.
func recalculateCmdTargetPoint(gameEventData s2prot.Struct) s2prot.Struct {

	targetPoint := gameEventData["TargetPoint"]
	if targetPoint != nil {
		targetPoint := gameEventData["TargetPoint"].(s2prot.Struct)

		nBits := 13
		divisor := 1 << nBits
		targetPointX := float64(targetPoint["x"].(int64)) / float64(divisor)
		targetPointY := float64(targetPoint["y"].(int64)) / float64(divisor)
		targetPointZ := float64(targetPoint["z"].(int64)) / float64(divisor)
		// targetPointZ := float64(targetPoint["z"].(int64))

		// REVIEW: sc2reader is not recalculating this coordinate,
		// at the same time scelight provides a recalculated float value:
		// - https://github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/cmd/TargetPoint.java#L41
		// - https://github.com/ggtracker/sc2reader/blob/ba8b52ec0875df5cd869af09dccdb4d9f84ae921/sc2reader/events/game.py#L267-L292

		targetPoint["x"] = targetPointX
		targetPoint["y"] = targetPointY
		targetPoint["z"] = targetPointZ
		gameEventData["TargetPoint"] = targetPoint
		return gameEventData
	}

	return gameEventData
}

// recalculateCmdSnapshotPoint recalculates the snapshotPoint coordinates
// for the targetUnit and returns the mutated gameEventData.
func recalculateCmdSnapshotPoint(gameEventData s2prot.Struct) s2prot.Struct {

	targetUnit := gameEventData["TargetUnit"]
	if targetUnit != nil {
		targetUnit := gameEventData["TargetUnit"].(s2prot.Struct)
		snapshotPoint := targetUnit["snapshotPoint"]
		if snapshotPoint != nil {
			snapshotPoint := targetUnit["snapshotPoint"].(s2prot.Struct)

			nBits := 13
			divisor := 1 << nBits
			snapshotPointX := float64(snapshotPoint["x"].(int64)) / float64(divisor)
			snapshotPointY := float64(snapshotPoint["y"].(int64)) / float64(divisor)
			snapshotPointZ := float64(snapshotPoint["z"].(int64)) / float64(divisor)

			snapshotPoint["x"] = snapshotPointX
			snapshotPoint["y"] = snapshotPointY
			snapshotPoint["z"] = snapshotPointZ
			targetUnit["snapshotPoint"] = snapshotPoint
		}
		gameEventData["TargetUnit"] = targetUnit
		return gameEventData
	}
	return gameEventData
}

// getCmdFlags translates the integer bitmask to a string array,
// and returns the mutated gameEventStruct.
func getCmdFlags(
	gameEvent s2prot.Event,
	gameEventStruct s2prot.Struct,
) s2prot.Struct {

	gameEventFlagBitmask := gameEvent.Int("cmdFlags")
	commandFlags := getCommandType(gameEventFlagBitmask)

	log.WithField("CommandType", commandFlags).Debug("Got command type")
	log.Debug("GameEventInt: ", gameEventFlagBitmask)

	gameEventStruct["cmdFlags"] = commandFlags
	return gameEventStruct
}

// getAbilityData acquires the ability data from the gameEventAbility struct.
// Mutates the gameEventAbility struct and returns it containing
// a human readable ability name and command name.
func getAbilityData(gameEventAbility s2prot.Struct) s2prot.Struct {

	// TODO: load this from balance data where possible
	// REVIEW: What to do if no balance data is available?
	// should this break the processing?
	// TODO: I have found the proper string ability name and the string ability command.
	// Either replace the values with strings,
	// maybe a better option is to add more fields such as abilityName and abilityCommandName
	gameEventAbility = getAbilityName(gameEventAbility)
	gameEventAbility = getAbilityCommandName(gameEventAbility)

	return gameEventAbility
}

// getAbilityName acquires the ability name, mutates the gameEventAbility and returns it.
func getAbilityName(gameEventAbility s2prot.Struct) s2prot.Struct {

	// TODO: This requires lookup to the game version balance data

	// 	gameEventAbility := gameEventAbility.(s2prot.Struct)
	// 	abilityLink := gameEventAbility["abilLink"].(int64)

	return gameEventAbility
}

// getAbilityCommandName acquires the ability specific command name,
// mutates the gameEventAbility and returns it.
func getAbilityCommandName(gameEventAbility s2prot.Struct) s2prot.Struct {

	// TODO: This requires lookup to the game version balance data

	// 	abilityCmdIndex := gameEventAbility["abilCmdIndex"].(int64)
	// 	abilCmdData := gameEventAbility["abilCmdData"]
	// 	if abilCmdData != nil {
	// 		log.Info("abilCmdData is not nil")
	// 	}
	// 	log.WithField("abilityLink", abilityLink).Debug("Got ability link")
	// 	log.WithField("abilityCmdIndex", abilityCmdIndex).Debug("Got ability cmdIndex")
	// 	log.WithField("abilCmdData", abilCmdData).Debug("Got ability cmdData")

	return gameEventAbility
}

func cleanCameraUpdateEvent(gameEventStruct s2prot.Struct) s2prot.Struct {

	// TODO: recalculate camera coordinates (if needed)
	// https: //github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/camera/TargetPoint.java#L41
	// https://github.com/icza/scelight/blob/7360c30765c9bc2f25b069da4377b37e47d4b426/src-app/hu/scelight/sc2/rep/model/gameevents/camera/CameraUpdateEvent.java#L52

	// REVIEW: Camera TargetPoint has different recalculation than Cmd TargetPoint
	recalculatePitchYaw(gameEventStruct)
	recalculateCameraTargetPoint(gameEventStruct)

	return gameEventStruct
}

func recalculatePitchYaw(gameEventData s2prot.Struct) s2prot.Struct {

	// Recalculate pitch to degrees
	pitch := gameEventData["pitch"]
	if pitch != nil {
		pitch := gameEventData["pitch"].(int64)
		// return pitch == null ? null : ( 45 * ( ( ( ( ( ( pitch << 5 ) - 0x2000 ) << 17 ) - 1 ) >> 17 ) + 1 ) ) / 4096.0f;
		recalculatedPitch := float64((45 * ((((((pitch << 5) - 0x2000) << 17) - 1) >> 17) + 1))) / 4096.0

		gameEventData["pitch"] = recalculatedPitch
	}

	// Recalculate yaw to degrees
	yaw := gameEventData["yaw"]
	if yaw != nil {
		yaw := gameEventData["yaw"].(int64)
		// return yaw == null ? null : ( 45 * ( ( ( ( ( ( yaw << 5 ) - 0x2000 ) << 17 ) - 1 ) >> 17 ) + 1 ) ) / 4096.0f;
		recalculatedYaw := float64((45 * ((((((yaw << 5) - 0x2000) << 17) - 1) >> 17) + 1))) / 4096.0
		gameEventData["yaw"] = recalculatedYaw
	}

	return gameEventData
}

func recalculateCameraTargetPoint(gameEventData s2prot.Struct) s2prot.Struct {

	targetPoint := gameEventData["target"]
	if targetPoint != nil {
		targetPoint := gameEventData["target"].(s2prot.Struct)

		targetPointX := float64(targetPoint["x"].(int64)) / 256.0
		targetPointY := float64(targetPoint["y"].(int64)) / 256.0

		targetPoint["x"] = targetPointX
		targetPoint["y"] = targetPointY

		// return distance == null ? null : distance / 256.0f;
		distance := targetPoint["distance"]
		if distance != nil {
			distance := float64(targetPoint["distance"].(int64)) / 256.0
			targetPoint["distance"] = distance
		}
	}

	gameEventData["target"] = targetPoint

	return gameEventData
}
