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

			// gameEventAbility := gameEventStruct["abil"]
			// if gameEventAbility != nil {
			// 	gameEventAbility := gameEventAbility.(s2prot.Struct)
			// 	abilityLink := gameEventAbility["abilLink"].(int64)
			// 	abilityCmdIndex := gameEventAbility["abilCmdIndex"].(int64)
			// 	abilCmdData := gameEventAbility["abilCmdData"]
			// 	if abilCmdData != nil {
			// 		log.Info("abilCmdData is not nil")
			// 	}
			// 	log.WithField("abilityLink", abilityLink).Debug("Got ability link")
			// 	log.WithField("abilityCmdIndex", abilityCmdIndex).Debug("Got ability cmdIndex")
			// 	log.WithField("abilCmdData", abilCmdData).Debug("Got ability cmdData")

			// TODO: load this from balance data where possible
			// REVIEW: What to do if no balance data is available?
			// should this break the processing?

			// TODO: I have found the proper string ability name and the string ability command.
			// Either replace the values with strings,
			// maybe a better option is to add more fields such as abilityName and abilityCommandName
			// }

			// gameEventData := gameEventStruct["data"]
			// if gameEventData != nil {
			// 	gameEventData := gameEventStruct["data"].(s2prot.Struct)

			// 	// Target point needs to be translated to the true coordinates:
			// 	targetPoint := gameEventData["targetPoint"]
			// 	// Scale x,y,z value to match the true coordinates:

			// 	if targetPoint != nil {
			// 		targetUnit := gameEventData["targetPoint"].(s2prot.Struct)
			// 		log.WithField("targetUnit", targetUnit).Debug("Got target unit")

			// 	}

			// TODO: SNAPSHOT POINTS SHOULD BE TRANSLATED TO TRUE CORRDINATES SEPARATELY:
			// snapshotPoint := gameEventData["snapshotPoint"].(s2prot.Struct)
			// snapshotPointX := float64(snapshotPoint["x"].(int64)) / 8192
			// snapshotPointY := float64(snapshotPoint["y"].(int64)) / 4096
			// snapshotPointZ := float64(snapshotPoint["z"].(int64)) / 4096

			// gameEventStruct["data"].(s2prot.Struct)["snapshotPoint"].(s2prot.Struct)["x"] = snapshotPointX
			// gameEventStruct["data"].(s2prot.Struct)["snapshotPoint"].(s2prot.Struct)["y"] = snapshotPointY
			// gameEventStruct["data"].(s2prot.Struct)["snapshotPoint"].(s2prot.Struct)["z"] = snapshotPointZ

			// }

			gameEventFlagBitmask := gameEvent.Int("cmdFlags")
			commandFlags := getCommandType(gameEventFlagBitmask)

			log.WithField("CommandType", commandFlags).Debug("Got command type")
			log.Debug("GameEventInt: ", gameEventFlagBitmask)

			gameEventStruct["cmdFlags"] = commandFlags
		}

		gameEventsStructs = append(gameEventsStructs, gameEventStruct)
	}
	log.Info("Defined cleanGameEvents struct")
	return gameEventsStructs
}
