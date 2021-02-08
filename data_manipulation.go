package main

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/icza/s2prot/rep"
)

// type CleanedReplay struct {
// 	header rep.Header
// }

func deleteUnusedObjects(replayData *rep.Rep) *rep.Rep {

	elapsedGameLoops := replayData.Header

	cleanHeader := data.CleanedHeader{}
	cleanInitData := data.CleanedInitData{}
	cleanDetails := data.CleanedDetails{}
	cleanMetadata := data.CleanedMetadata{}

	dirtyMessageEvents := replayData.MessageEvts
	dirtyGameEvents := replayData.GameEvts
	dirtyTrackerEvents := replayData.TrackerEvts.Evts
	dirtyPIDPlayerDescMap := replayData.TrackerEvts.PIDPlayerDescMap
	dirtyToonPlayerDescMap := replayData.TrackerEvts.ToonPlayerDescMap
	justGameEvtsErr := replayData.GameEvtsErr

	justMessageEvtsErr := replayData.MessageEvtsErr
	justTrackerEvtsErr := replayData.TrackerEvtsErr

	cleanedReplay := data.CleanedReplay{
		Header:            cleanHeader,
		InitData:          cleanInitData,
		Details:           cleanDetails,
		Metadata:          cleanMetadata,
		MessageEvents:     dirtyMessageEvents,
		GameEvents:        dirtyGameEvents,
		TrackerEvents:     dirtyTrackerEvents,
		PIDPlayerDescMap:  dirtyPIDPlayerDescMap,
		ToonPlayerDescMap: dirtyToonPlayerDescMap,
		GameEvtsErr:       justGameEvtsErr,
		MessageEvtsErr:    justMessageEvtsErr,
		TrackerEvtsErr:    justTrackerEvtsErr,
	}

	// TODO: Initialize structs defined in custom_types directory

	// TODO: Define for loops that will be checking different event types and not creating instances if the event type is unwanted
	// Good example of that will be some of the chat events that are in messageEvents.

	// TODO: Initialize my own type of CleanedReplay only with the fields that are needed.

	return replayData
}

func anonymizeReplayData(replayData *rep.Rep) *rep.Rep {

	// TODO: Anonymize the information about players.
	// This needs to be done by calling some external file and / or memory which will be holding persistent information about all of the players.

	return replayData
}
