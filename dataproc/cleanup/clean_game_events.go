package cleanup

import (
	"encoding/json"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/cleanup/game_events"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// CleanGameEvents copies the game events,
// has the capability of removing unecessary fields.
func CleanGameEvents(replayData *rep.Rep) []map[string]any {
	log.Debug("Entered CleanGameEvents()")

	// Constructing a clean GameEvents without unescessary fields:
	// NOTE: cleanGameEvents are of type map[string]any because to effectively
	// add, recalculate or remove field from the game events it is easier to work with
	// maps than with s2prot.Structs, maps are related closer to the final JSON format.
	var cleanGameEvents []map[string]any
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
func cleanGameEvent(gameEvent s2prot.Event) (map[string]any, error) {
	log.Debug("Entered cleanGameEvent()")

	gameEventBytes := gameEvent.Struct.String()
	var gameEventJSONMap map[string]any
	err := json.Unmarshal([]byte(gameEventBytes), &gameEventJSONMap)
	if err != nil {
		log.Error("Failed to unmarshal the s2prot.Struct from GameEvents")
		return nil, err
	}

	// REVIEW: Verify if there are any more game event types that need to be cleaned:
	// recalculateGameEventTarget(gameEventJSONMap)
	switch gameEvent.Name {
	case "Cmd":
		game_events.CleanCmdEvent(gameEvent, gameEventJSONMap)
	case "CameraUpdate":
		game_events.CleanCameraUpdateEvent(gameEventJSONMap)
	case "CameraSave":
		game_events.CleanCameraSaveEvent(gameEventJSONMap)
	case "CmdUpdateTargetUnit":
		game_events.CleanCmdUpdateTargetUnitEvent(gameEventJSONMap)
	case "CmdUpdateTargetPoint":
		game_events.CleanCmdUpdateTargetPointEvent(gameEventJSONMap)
	}

	log.Debug("Finished cleanGameEvent()")
	return gameEventJSONMap, nil

}
