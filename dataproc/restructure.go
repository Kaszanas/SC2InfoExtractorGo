package dataproc

import (
	"encoding/json"
	"time"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func redifineReplayStructure(replayData *rep.Rep) (data.CleanedReplay, bool) {

	// TODO: Move user initData to playerList

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

	gameSpeed := gameDescription.Struct["gameSpeed"].(int64)
	gameSpeedChecked, okGameSpeed := checkUint8(gameSpeed)
	if !okGameSpeed {
		log.Error("Found that the value of gameSpeed exceeds uint8")
		return data.CleanedReplay{}, false
	}

	isBlizzardMap := gameDescription.Struct["isBlizzardMap"].(bool)
	mapAuthorName := gameDescription.Struct["mapAuthorName"].(string)

	mapFileSyncChecksum := gameDescription.Struct["mapFileSyncChecksum"].(int64)

	mapSizeX := gameDescription.Struct["mapSizeX"].(int64)
	mapSizeXChecked, okMapSizeX := checkUint32(mapSizeX)
	if !okMapSizeX {
		log.WithField("mapSizeX", mapSizeX).Error("Found that value of mapSizeX exceeds uint32")
		return data.CleanedReplay{}, false
	}

	mapSizeY := gameDescription.Struct["mapSizeY"].(int64)
	mapSizeYChecked, okMapSizeY := checkUint32(mapSizeY)
	if !okMapSizeY {
		log.WithField("mapSizeY", mapSizeY).Error("Found that value of mapSizeY exceeds uint32")
		return data.CleanedReplay{}, false
	}

	maxPlayers := gameDescription.Struct["maxPlayers"].(int64)
	maxPlayersChecked, okMaxPlayers := checkUint8(maxPlayers)
	if !okMaxPlayers {
		log.WithField("maxPlayers", maxPlayers).Error("Found that value of maxPlayers exceeds uint8")
		return data.CleanedReplay{}, false
	}

	cleanedGameDescription := data.CleanedGameDescription{
		GameOptions:         gameOptions,
		GameSpeed:           gameSpeedChecked,
		IsBlizzardMap:       isBlizzardMap,
		MapAuthorName:       mapAuthorName,
		MapFileSyncChecksum: mapFileSyncChecksum,
		MapSizeX:            mapSizeXChecked,
		MapSizeY:            mapSizeYChecked,
		MaxPlayers:          maxPlayersChecked,
	}

	// Constructing a clean UserInitData without unescessary fields:
	var cleanedUserInitDataList []data.CleanedUserInitData
	for _, userInitData := range replayData.InitData.UserInitDatas {
		// If the name is an empty string ommit the struct and enter next iteration:
		name := userInitData.Name()
		if !(len(name) > 0) {
			continue
		}

		combinedRaceLevels := userInitData.CombinedRaceLevels()
		combinedRaceLevelsChecked, okCombinedRaceLevels := checkUint64(combinedRaceLevels)
		if !okCombinedRaceLevels {
			log.WithField("combinedRaceLevels", combinedRaceLevels).Error("Found that value of combinedRaceLevels exceeds uint64")
			return data.CleanedReplay{}, false
		}

		highestLeague := userInitData.Struct["highestLeague"].(int64)
		highestLeagueChecked, okHighestLeague := checkUint32(highestLeague)
		if !okHighestLeague {
			log.WithField("highestLeague", highestLeague).Error("Found that value of highestLeague exceeds uint32")
			return data.CleanedReplay{}, false
		}

		clanTag := userInitData.Struct["clanTag"].(string)
		isInClan := checkClan(clanTag)

		userInitDataStruct := data.CleanedUserInitData{
			CombinedRaceLevels: combinedRaceLevelsChecked,
			HighestLeague:      highestLeagueChecked,
			Name:               name,
			IsInClan:           isInClan,
		}

		cleanedUserInitDataList = append(cleanedUserInitDataList, userInitDataStruct)
	}

	cleanInitData := data.CleanedInitData{
		GameDescription: cleanedGameDescription,
	}

	// Constructing a clean CleanedDetails without unescessary fields
	details := replayData.Details
	detailsGameSpeed := uint8(details.Struct["gameSpeed"].(int64))
	detailsIsBlizzardMap := details.IsBlizzardMap()

	var detailsPlayerList []data.CleanedPlayerListStruct
	for _, initPlayer := range cleanedUserInitDataList {

		for _, player := range details.Players() {
			if initPlayer.Name == player.Name {

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

				result := player.Struct["result"].(int64)
				resultChecked, okResult := checkUint8(result)
				if !okResult {
					log.WithField("result", result).Error("Found that value of result exceeds uint8")
					return data.CleanedReplay{}, false
				}

				teamID := player.TeamID()
				teamIDChecked, okTeamID := checkUint8(teamID)
				if !okTeamID {
					log.WithField("teamID", teamID).Error("Found that value of result exceeds uint8")
					return data.CleanedReplay{}, false
				}

				// TODO: Check if this cannot be done easier?
				// Accessing toon data by Golang magic:
				toon := player.Struct["toon"]
				intermediateJSON, err := json.Marshal(&toon)
				if err != nil {
					log.WithField("error", err).Error("Encountered error while json marshaling")
					return data.CleanedReplay{}, false
				}
				var unmarshalledData interface{}
				err = json.Unmarshal(intermediateJSON, &unmarshalledData)

				if err != nil {
					log.WithField("error", err).Error("Encountered error while json unmarshaling")
					return data.CleanedReplay{}, false
				}
				toonMap := unmarshalledData.(map[string]interface{})

				regionChecked, regionOk := checkUint8Float(toonMap["region"].(float64))
				if !regionOk {
					log.Error("Found that value of region exceeds uint8")
					return data.CleanedReplay{}, false
				}

				realmChecked, realmOk := checkUint8Float(toonMap["realm"].(float64))
				if !realmOk {
					log.Error("Found that value of realm exceeds uint8")
					return data.CleanedReplay{}, false
				}

				// Checking the region and realm strings for the players:
				regionString := rep.Regions[int(regionChecked)].String()
				realmString := replayData.InitData.GameDescription.Region().Realms[int(realmChecked)].String()

				cleanedPlayerStruct := data.CleanedPlayerListStruct{
					Name:               name,
					Race:               race,
					Result:             resultChecked,
					IsInClan:           initPlayer.IsInClan,
					HighestLeague:      initPlayer.HighestLeague,
					Handicap:           handicap,
					TeamID:             teamIDChecked,
					Realm:              regionString,
					Region:             realmString,
					CombinedRaceLevels: initPlayer.CombinedRaceLevels,
					Color:              playerColor,
				}

				detailsPlayerList = append(detailsPlayerList, cleanedPlayerStruct)

			}
		}
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
		ToonPlayerDescMap: dirtyToonPlayerDescMap,
		GameEvtsErr:       justGameEvtsErr,
		MessageEvtsErr:    justMessageEvtsErr,
		TrackerEvtsErr:    justTrackerEvtsErr,
	}

	return cleanedReplay, true
}