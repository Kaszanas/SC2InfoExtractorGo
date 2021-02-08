package main

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

// type CleanedReplay struct {
// 	header rep.Header
// }

func deleteUnusedObjects(replayData *rep.Rep) *rep.Rep {

	// Constructing a clean replay header without unescessary fields:
	elapsedGameLoops := replayData.Header.Struct["elapsedGameLoops"].(int64)
	duration := replayData.Header.Duration()
	useScaledTime := replayData.Header.Struct["useScaledTime"].(bool)
	version := replayData.Header.Struct["version"].(s2prot.Struct)

	cleanHeader := data.CleanedHeader{
		ElapsedGameLoops: uint64(elapsedGameLoops),
		Duration:         duration,
		UseScaledTime:    useScaledTime,
		Version:          version,
	}

	// Constructing a clean GameDescription without unescessary fields:
	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions.Struct
	gameSpeed := uint8(gameDescription.Struct["gameSpeed"].(int64))
	isBlizzardMap := gameDescription.Struct["isBlizzardMap"].(bool)
	mapAuthorName := gameDescription.Struct["mapAuthorName"].(string)
	mapFileSyncChecksum := gameDescription.Struct["mapFileSyncChecksum"].(int)
	mapSizeX := uint32(gameDescription.Struct["mapSizeX"].(int))
	mapSizeY := uint32(gameDescription.Struct["mapSizeY"].(int))
	maxPlayers := uint8(gameDescription.Struct["maxPlayers"].(int))

	cleanedGameDescription := data.CleanedGameDescription{
		GameOptions:         gameOptions,
		GameSpeed:           gameSpeed,
		IsBlizzardMap:       isBlizzardMap,
		MapAuthorName:       mapAuthorName,
		MapFileSyncChecksum: mapFileSyncChecksum,
		MapSizeX:            mapSizeX,
		MapSizeY:            mapSizeY,
		MaxPlayers:          maxPlayers,
	}

	// Constructing a clean UserInitData without unescessary fields:
	// TODO: Iterate over user initial datas using a loop and construct my own types:
	combinedRaceLevels := replayData.InitData.UserInitDatas

	cleanUserInitData := data.CleanedUserInitData{}

	cleanInitData := data.CleanedInitData{GameDescription: cleanedGameDescription}
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
