package game_events

import (
	"github.com/icza/s2prot"
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

// CleanCmdEvent cleans the Cmd game event. It acuires the ability data,
// recalculates the snapshotPoint for a TargetUnit and the TargetPoint.
// Finally, it acquire the command flags and mutate the gameEventJSONMap.
func CleanCmdEvent(
	gameEvent s2prot.Event,
	gameEventJSONMap map[string]any,
) {

	log.Debug("Entered cleanCmdEvent()")

	if gameEventAbility, ok := gameEventJSONMap["abil"]; ok {
		if gameEventAbility == nil {
			log.Debug("Detected nil game event ability")
		} else {
			castedGameEventAbility := gameEventAbility.(map[string]any)
			// Acquiring ability data:
			getAbilityData(castedGameEventAbility)
			gameEventJSONMap["abil"] = castedGameEventAbility
		}
	}

	// Recalculating snapshotPoint for a TargetUnit:
	if gameEventData, ok := gameEventJSONMap["data"]; ok {
		if gameEventData == nil {
			log.Debug("Detected nil game event data")
		} else {
			castedGameEventData := gameEventData.(map[string]any)

			// Recalculating SnapshotPoint for a TargetUnit:
			recalculateCmdTargetUnitSnapshotPoint(castedGameEventData)

			// Recalculating TargetPoint:
			recalculateCmdTargetPoint(castedGameEventData)

			gameEventJSONMap["data"] = castedGameEventData
		}
	}

	// Getting command flags:
	getCmdFlags(gameEvent, gameEventJSONMap)

	log.Debug("Finished cleanCmdEvent()")
}

// getCmdFlags translates the integer bitmask to a string array,
// and returns the mutated gameEventStruct.
func getCmdFlags(
	gameEvent s2prot.Event,
	gameEventJSONMap map[string]any,
) {

	log.Debug("Entered getCmdFlags()")

	gameEventFlagBitmask := gameEvent.Int("cmdFlags")
	commandFlags := getCommandType(gameEventFlagBitmask)

	log.WithField("CommandType", commandFlags).Debug("Got command type")
	log.Debug("GameEventInt: ", gameEventFlagBitmask)

	gameEventJSONMap["cmdFlags"] = commandFlags

	log.Debug("Finished getCmdFlags()")
}

// getCommandType translates the integer bitmask to a string array containing the command type.
// Returns the string array.
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
func getAbilityData(gameEventJSONMap map[string]any) {

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
	gameEventJSONMap map[string]any,
) map[string]any {

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
	gameEventJSONMap map[string]any,
) map[string]any {

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
