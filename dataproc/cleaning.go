package dataproc

import (
	"encoding/json"
	"time"

	"github.com/Kaszanas/GoSC2Science/datastruct"
	data "github.com/Kaszanas/GoSC2Science/datastruct"
	settings "github.com/Kaszanas/GoSC2Science/settings"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

// TODO: Rename this file to restructuring.go because this is not cleaning the data it is just ommiting fields which for sure will not be used.

// TODO: Introduce logging.
func checkClan(clanTag string) bool {
	if clanTag != "" {
		return true
	} else {
		return false
	}
}

func redifineReplayStructure(replayData *rep.Rep) (data.CleanedReplay, bool) {

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
	mapFileSyncChecksum := gameDescription.Struct["mapFileSyncChecksum"].(int64)
	mapSizeX := uint32(gameDescription.Struct["mapSizeX"].(int64))
	mapSizeY := uint32(gameDescription.Struct["mapSizeY"].(int64))
	maxPlayers := uint8(gameDescription.Struct["maxPlayers"].(int64))

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
	var cleanedUserInitDataList []data.CleanedUserInitData
	for _, userInitData := range replayData.InitData.UserInitDatas {
		combinedRaceLevels := uint64(userInitData.CombinedRaceLevels())
		highestLeague := uint32(userInitData.Struct["highestLeague"].(int64))
		name := userInitData.Name()
		clanTag := userInitData.Struct["clanTag"].(string)
		isInClan := checkClan(clanTag)

		userInitDataStruct := data.CleanedUserInitData{
			CombinedRaceLevels: combinedRaceLevels,
			HighestLeague:      highestLeague,
			Name:               name,
			IsInClan:           isInClan,
		}

		cleanedUserInitDataList = append(cleanedUserInitDataList, userInitDataStruct)
	}

	cleanInitData := data.CleanedInitData{
		GameDescription: cleanedGameDescription,
		UserInitData:    cleanedUserInitDataList,
	}

	// Constructing a clean CleanedDetails without unescessary fields
	details := replayData.Details
	detailsGameSpeed := uint8(details.Struct["gameSpeed"].(int64))
	detailsIsBlizzardMap := details.IsBlizzardMap()

	var detailsPlayerList []data.CleanedPlayerListStruct
	for _, player := range details.Players() {
		colorA := uint8(player.Color[0])
		colorB := uint8(player.Color[1])
		colorG := uint8(player.Color[2])
		colorR := uint8(player.Color[3])
		playerColor := data.PlayerListColor{
			A: colorA,
			B: colorB,
			G: colorG,
			R: colorR,
		}

		handicap := uint8(player.Handicap())
		name := player.Name
		race := player.Struct["race"].(string)
		result := uint8(player.Struct["result"].(int64))
		teamID := uint8(player.TeamID())

		// Accessing toon data by Golang magic:
		toon := player.Struct["toon"]
		intermediateJSON, err := json.Marshal(&toon)

		// TODO: Error logging and handling
		if err != nil {
			return data.CleanedReplay{}, false
		}
		var unmarshalledData interface{}
		err = json.Unmarshal(intermediateJSON, &unmarshalledData)

		// TODO: Error logging and handling
		if err != nil {
			return data.CleanedReplay{}, false
		}
		toonMap := unmarshalledData.(map[string]interface{})

		realm := uint8(toonMap["realm"].(float64))
		region := uint8(toonMap["region"].(float64))

		cleanedPlayerStruct := data.CleanedPlayerListStruct{
			Color:    playerColor,
			Handicap: handicap,
			Name:     name,
			Race:     race,
			Result:   result,
			TeamID:   teamID,
			Realm:    realm,
			Region:   region,
		}

		detailsPlayerList = append(detailsPlayerList, cleanedPlayerStruct)
	}

	timeLocalOffset := details.TimeLocalOffset()
	timeUTC := details.TimeUTC()
	mapName := details.Struct["title"].(string)

	cleanDetails := data.CleanedDetails{
		GameSpeed:       detailsGameSpeed,
		IsBlizzardMap:   detailsIsBlizzardMap,
		PlayerList:      detailsPlayerList,
		TimeLocalOffset: timeLocalOffset,
		TimeUTC:         timeUTC,
		MapName:         mapName,
	}

	// Constructing a clean CleanedMetadata without unescessary fields:
	metadata := replayData.Metadata
	metadataBaseBuild := metadata.BaseBuild()
	metadataDataBuild := metadata.DataBuild()
	metadataDuration := time.Duration(metadata.Struct["Duration"].(float64))
	metadataGameVersion := metadata.GameVersion()

	var metadataCleanedPlayersList []data.CleanedPlayer
	for _, player := range metadata.Players() {

		playerID := uint8(player.PlayerID())
		apm := uint16(player.APM())
		mmr := uint16(player.MMR())
		result := player.Result()
		assignedRace := player.AssignedRace()
		selectedRace := player.SelectedRace()

		cleanedPlayerStruct := data.CleanedPlayer{
			PlayerID:     playerID,
			APM:          apm,
			MMR:          mmr,
			Result:       result,
			AssignedRace: assignedRace,
			SelectedRace: selectedRace,
		}
		metadataCleanedPlayersList = append(metadataCleanedPlayersList, cleanedPlayerStruct)
	}

	metadataMapName := metadata.Title()

	cleanMetadata := data.CleanedMetadata{
		BaseBuild:   metadataBaseBuild,
		DataBuild:   metadataDataBuild,
		Duration:    metadataDuration,
		GameVersion: metadataGameVersion,
		Players:     metadataCleanedPlayersList,
		MapName:     metadataMapName,
	}

	dirtyMessageEvents := replayData.MessageEvts
	dirtyGameEvents := replayData.GameEvts
	dirtyTrackerEvents := replayData.TrackerEvts.Evts
	dirtyPIDPlayerDescMap := replayData.TrackerEvts.PIDPlayerDescMap
	dirtyToonPlayerDescMap := replayData.TrackerEvts.ToonPlayerDescMap
	justGameEvtsErr := replayData.GameEvtsErr

	var messageEventsStructs []s2prot.Struct
	for _, messageEvent := range dirtyMessageEvents {
		messageEventsStructs = append(messageEventsStructs, messageEvent.Struct)
	}

	var gameEventsStructs []s2prot.Struct
	for _, gameEvent := range dirtyGameEvents {
		gameEventsStructs = append(gameEventsStructs, gameEvent.Struct)
	}

	var trackerEventsStructs []s2prot.Struct
	for _, trackerEvent := range dirtyTrackerEvents {
		trackerEventsStructs = append(trackerEventsStructs, trackerEvent.Struct)
	}

	justMessageEvtsErr := replayData.MessageEvtsErr
	justTrackerEvtsErr := replayData.TrackerEvtsErr

	cleanedReplay := data.CleanedReplay{
		Header:            cleanHeader,
		InitData:          cleanInitData,
		Details:           cleanDetails,
		Metadata:          cleanMetadata,
		MessageEvents:     messageEventsStructs,
		GameEvents:        gameEventsStructs,
		TrackerEvents:     trackerEventsStructs,
		PIDPlayerDescMap:  dirtyPIDPlayerDescMap,
		ToonPlayerDescMap: dirtyToonPlayerDescMap,
		GameEvtsErr:       justGameEvtsErr,
		MessageEvtsErr:    justMessageEvtsErr,
		TrackerEvtsErr:    justTrackerEvtsErr,
	}

	return cleanedReplay, true
}

// Helper function checking if a slice contains a string.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func cleanReplayStructure(replayData *datastruct.CleanedReplay) bool {

	var anonymizedMessageEvents []s2prot.Struct
	for _, event := range replayData.MessageEvents {
		eventType := event["evtTypeName"].(string)
		if !contains(settings.UnusedMessageEvents(), eventType) {
			anonymizedMessageEvents = append(anonymizedMessageEvents, event)
		}
	}

	var anonymizedGameEvents []s2prot.Struct
	for _, event := range replayData.GameEvents {
		if !contains(settings.UnusedGameEvents(), event["evtTypeName"].(string)) {
			anonymizedGameEvents = append(anonymizedGameEvents, event)
		}
	}

	replayData.MessageEvents = anonymizedMessageEvents
	replayData.GameEvents = anonymizedGameEvents

	return true

}
