package dataproc

import (
	"strings"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// redefineReplayStructure moves arbitrary data into different data structures.
func redifineReplayStructure(
	replayData *rep.Rep,
	englishToForeignMapping map[string]string,
) (replay_data.CleanedReplay, bool) {

	log.Info("Entered redefineReplayStructure()")

	cleanHeader := cleanHeader(replayData)
	cleanGameDescription, ok := cleanGameDescription(replayData)
	if !ok {
		return replay_data.CleanedReplay{}, false
	}
	cleanInitData, cleanedUserInitDataList, ok := cleanInitData(
		replayData,
		cleanGameDescription)
	if !ok {
		return replay_data.CleanedReplay{}, false
	}
	cleanDetails, detailsReplayMapField := cleanDetails(replayData)
	cleanMetadata, metadataReplayMapField := cleanMetadata(replayData)

	mapFields := []replay_data.ReplayMapField{
		detailsReplayMapField,
		metadataReplayMapField,
	}
	ok = adjustMapName(mapFields, englishToForeignMapping, &cleanMetadata)
	if !ok {
		log.Error("Failed to adjust map name!")
		return replay_data.CleanedReplay{}, false
	}

	enhancedToonDescMap := cleanToonDescMap(replayData, cleanedUserInitDataList)

	messageEventsStructs := cleanMessageEvents(replayData)
	gameEventsStructs := cleanGameEvents(replayData)
	trackerEventsStructs := cleanTrackerEvents(replayData)

	justMessageEvtsErr := replayData.MessageEvtsErr
	justTrackerEvtsErr := replayData.TrackerEvtsErr
	justGameEvtsErr := replayData.GameEvtsErr
	cleanedReplay := replay_data.CleanedReplay{
		Header:            cleanHeader,
		InitData:          cleanInitData,
		Details:           cleanDetails,
		Metadata:          cleanMetadata,
		MessageEvents:     messageEventsStructs,
		GameEvents:        gameEventsStructs,
		TrackerEvents:     trackerEventsStructs,
		ToonPlayerDescMap: enhancedToonDescMap,
		GameEvtsErr:       justGameEvtsErr,
		MessageEvtsErr:    justMessageEvtsErr,
		TrackerEvtsErr:    justTrackerEvtsErr,
	}
	log.Info("Defined cleanedReplay struct")

	log.Info("Finished cleanReplayStructure()")

	return cleanedReplay, true
}

// adjustMapName takes multiple map fields, finds the first non-empty one
// and adjusts the map name in CleanedMetadata with the version available
// in englishToForeignMapping.
func adjustMapName(
	mapFields []replay_data.ReplayMapField,
	englishToForeignMapping map[string]string,
	cleanMetadata *replay_data.CleanedMetadata,
) bool {

	// Got map name from metadata and details, searching for the first non-empty one:
	foreignMapName := replay_data.CombineReplayMapFields(mapFields)
	if foreignMapName == "" {
		log.Error("Failed to combine map name!")
		return false
	}
	// Attempting to acquire the english map name:
	englishMapName, ok := englishToForeignMapping[foreignMapName]
	if !ok {
		log.WithField("foreignMapName", foreignMapName).
			Error("Map name not found in englishToForeignMapping!")
		return false
	}

	// Adjusting the map name in CleanedMetadata:
	cleanMetadata.MapName = englishMapName

	return true
}

// cleanHeader copies the header,
// has the capability of removing unescessary fields.
func cleanHeader(replayData *rep.Rep) replay_data.CleanedHeader {
	// Constructing a clean replay header without unescessary fields:
	elapsedGameLoops := replayData.Header.Loops()
	// TODO: These values of duration are not verified: https://github.com/icza/s2prot/issues/48
	// durationNanoseconds := replayData.Header.Duration().Nanoseconds()
	// durationSeconds := replayData.Header.Duration().Seconds()
	// version := replayData.Header.Struct["version"].(s2prot.Struct)

	version := replayData.Header.VersionString()

	cleanHeader := replay_data.CleanedHeader{
		ElapsedGameLoops: uint64(elapsedGameLoops),
		// DurationNanoseconds: durationNanoseconds,
		// DurationSeconds:     durationSeconds,
		Version: version,
	}
	log.Info("Defined cleanHeader struct")
	return cleanHeader
}

// cleanGameDescription copies the game description,
// partly verifies the integrity of the data.
// Has the capability to remove unescessary fields.
func cleanGameDescription(replayData *rep.Rep) (replay_data.CleanedGameDescription, bool) {

	// Constructing a clean GameDescription without unescessary fields:
	gameDescription := replayData.InitData.GameDescription
	gameOptions := gameDescription.GameOptions.Struct

	gameSpeedString := gameDescription.GameSpeed().String()

	isBlizzardMap := gameDescription.IsBlizzardMap()
	mapAuthorName := gameDescription.MapAuthorName()

	// mapFilename := gameDescription.MapFileName()
	// log.WithField("mapFilename", mapFilename).Info("Found mapFilename")

	mapFileSyncChecksum := gameDescription.MapFileSyncChecksum()

	mapSizeX := gameDescription.MapSizeX()
	mapSizeXChecked, okMapSizeX := checkUint32(mapSizeX)
	if !okMapSizeX {
		log.WithField("mapSizeX", mapSizeX).
			Error("Found that value of mapSizeX exceeds uint32")
		return replay_data.CleanedGameDescription{}, false
	}

	mapSizeY := gameDescription.MapSizeY()
	mapSizeYChecked, okMapSizeY := checkUint32(mapSizeY)
	if !okMapSizeY {
		log.WithField("mapSizeY", mapSizeY).
			Error("Found that value of mapSizeY exceeds uint32")
		return replay_data.CleanedGameDescription{}, false
	}

	maxPlayers := gameDescription.MaxPlayers()
	maxPlayersChecked, okMaxPlayers := checkUint8(maxPlayers)
	if !okMaxPlayers {
		log.WithField("maxPlayers", maxPlayers).
			Error("Found that value of maxPlayers exceeds uint8")
		return replay_data.CleanedGameDescription{}, false
	}

	cleanedGameDescription := replay_data.CleanedGameDescription{
		GameOptions:         gameOptions,
		GameSpeed:           gameSpeedString,
		IsBlizzardMap:       isBlizzardMap,
		MapAuthorName:       mapAuthorName,
		MapFileSyncChecksum: mapFileSyncChecksum,
		MapSizeX:            mapSizeXChecked,
		MapSizeY:            mapSizeYChecked,
		MaxPlayers:          maxPlayersChecked,
	}
	log.Info("Defined cleanedGameDescription struct")

	return cleanedGameDescription, true
}

// cleanInitData copies the init data,
// partly verifies the integrity of the data.
// Has the capability to remove unescessary fields.
func cleanInitData(
	replayData *rep.Rep,
	cleanedGameDescription replay_data.CleanedGameDescription) (
	replay_data.CleanedInitData,
	[]replay_data.CleanedUserInitData,
	bool) {
	// Constructing a clean UserInitData without unescessary fields:
	var cleanedUserInitDataList []replay_data.CleanedUserInitData
	for _, userInitData := range replayData.InitData.UserInitDatas {
		// If the name is an empty string ommit the struct and enter next iteration:
		name := userInitData.Name()
		if !(len(name) > 0) {
			continue
		}

		combinedRaceLevels := userInitData.CombinedRaceLevels()
		highestLeague := userInitData.HighestLeague().String()
		clanTag := userInitData.ClanTag()
		isInClan := checkClan(clanTag)

		userInitDataStruct := replay_data.CleanedUserInitData{
			CombinedRaceLevels: uint64(combinedRaceLevels),
			HighestLeague:      highestLeague,
			Name:               name,
			IsInClan:           isInClan,
			ClanTag:            clanTag,
		}

		cleanedUserInitDataList = append(
			cleanedUserInitDataList,
			userInitDataStruct,
		)
	}

	cleanInitData := replay_data.CleanedInitData{
		GameDescription: cleanedGameDescription,
	}
	log.Info("Defined cleanInitData struct")
	return cleanInitData, cleanedUserInitDataList, true
}

// cleanDetails copies the details,
// has the capability of removing unescessary fields.
func cleanDetails(replayData *rep.Rep) (replay_data.CleanedDetails, replay_data.ReplayMapField) {
	// Constructing a clean CleanedDetails without unescessary fields
	detailsGameSpeed := replayData.Details.GameSpeed().String()
	detailsIsBlizzardMap := replayData.Details.IsBlizzardMap()

	// mapFileName := replayData.Details.MapFileName()
	// log.WithField("mapFileName", mapFileName).Info("Found mapFileName")

	timeUTC := replayData.Details.TimeUTC()
	mapNameString := replayData.Details.Title()
	replayMapField := replay_data.ReplayMapField{
		MapName: mapNameString,
	}

	cleanDetails := replay_data.CleanedDetails{
		GameSpeed:     detailsGameSpeed,
		IsBlizzardMap: detailsIsBlizzardMap,
		// PlayerList:    detailsPlayerList, // Information from that part is merged with ToonDescMap
		// TimeLocalOffset: timeLocalOffset, // This is unused
		TimeUTC: timeUTC,
		// MapName: mapNameString, // This is unused
	}
	log.Info("Defined cleanDetails struct")
	return cleanDetails, replayMapField
}

// cleanMetadata copies the metadata,
// has the capability of removing unescessary fields.
func cleanMetadata(
	replayData *rep.Rep,
) (replay_data.CleanedMetadata, replay_data.ReplayMapField) {
	// Constructing a clean CleanedMetadata without unescessary fields:
	metadataBaseBuild := replayData.Metadata.BaseBuild()
	metadataDataBuild := replayData.Metadata.DataBuild()
	// metadataDuration := replayData.Metadata.DurationSec()
	metadataGameVersion := replayData.Metadata.GameVersion()

	foreignMetadataMapName := replayData.Metadata.Title()
	mapNameField := replay_data.ReplayMapField{
		MapName: foreignMetadataMapName,
	}

	cleanMetadata := replay_data.CleanedMetadata{
		BaseBuild: metadataBaseBuild,
		DataBuild: metadataDataBuild,
		// Duration:    metadataDuration,
		GameVersion: metadataGameVersion,
		// Players:     metadataCleanedPlayersList, // This is unused.
		MapName: foreignMetadataMapName,
	}
	log.Info("Defined cleanMetadata struct")
	return cleanMetadata, mapNameField
}

// cleanToonDescMap copies the toon description map,
// changes the structure into a more readable form.
func cleanToonDescMap(
	replayData *rep.Rep,
	cleanedUserInitDataList []replay_data.CleanedUserInitData) map[string]replay_data.EnhancedToonDescMap {

	dirtyToonPlayerDescMap := replayData.TrackerEvts.ToonPlayerDescMap

	// Merging data-structures to data.EnhancedToonDescMap
	enhancedToonDescMap := make(map[string]replay_data.EnhancedToonDescMap)
	for toonKey, playerDescription := range dirtyToonPlayerDescMap {
		var initializedToonDescMap replay_data.EnhancedToonDescMap
		enhancedToonDescMap[toonKey] = initializedToonDescMap

		// Merging information held in metadata.Players into data.EnhancedToonDescMap
		for _, player := range replayData.Metadata.Players() {
			if player.PlayerID() == playerDescription.PlayerID {
				metadataToonDescMap := enhancedToonDescMap[toonKey]
				// Filling out struct fields:
				metadataToonDescMap.PlayerID = playerDescription.PlayerID
				metadataToonDescMap.UserID = playerDescription.UserID
				metadataToonDescMap.SQ = playerDescription.SQ
				metadataToonDescMap.SupplyCappedPercent = playerDescription.SupplyCappedPercent
				metadataToonDescMap.StartDir = playerDescription.StartDir
				metadataToonDescMap.StartLocX = playerDescription.StartLocX
				metadataToonDescMap.StartLocY = playerDescription.StartLocY
				metadataToonDescMap.AssignedRace = player.AssignedRace()
				metadataToonDescMap.SelectedRace = player.SelectedRace()
				metadataToonDescMap.APM = player.APM()
				metadataToonDescMap.MMR = player.MMR()
				metadataToonDescMap.Result = player.Result()
				enhancedToonDescMap[toonKey] = metadataToonDescMap
			}
		}

		// Merging information contained in the details part of the replay:
		for _, player := range replayData.Details.Players() {
			if player.Toon.String() == toonKey {
				detailsEnhancedToonDescMap := enhancedToonDescMap[toonKey]

				detailsEnhancedToonDescMap.Name = player.Name

				// Checking if previously ran loop populated the Race information
				if detailsEnhancedToonDescMap.AssignedRace == "" {
					raceLetter := player.Race().Letter
					if raceLetter == 'T' {
						detailsEnhancedToonDescMap.AssignedRace = "Terr"
					}
					if raceLetter == 'P' {
						detailsEnhancedToonDescMap.AssignedRace = "Prot"
					}
					if raceLetter == 'Z' {
						detailsEnhancedToonDescMap.AssignedRace = "Zerg"
					}
				}
				if detailsEnhancedToonDescMap.Result == "" {
					detailsEnhancedToonDescMap.Result = player.Result().String()
				}

				// Filling out struct fields:
				detailsEnhancedToonDescMap.Region = player.Toon.Region().Name
				detailsEnhancedToonDescMap.Realm = player.Toon.Realm().Name
				detailsEnhancedToonDescMap.Color.A = player.Color[0]
				detailsEnhancedToonDescMap.Color.B = player.Color[1]
				detailsEnhancedToonDescMap.Color.G = player.Color[2]
				detailsEnhancedToonDescMap.Color.R = player.Color[3]
				detailsEnhancedToonDescMap.Handicap = player.Handicap()
				enhancedToonDescMap[toonKey] = detailsEnhancedToonDescMap
			}

			// Merging cleanedUserInitDataList information into data.EnhancedToonDescMap:
			for _, initPlayer := range cleanedUserInitDataList {
				if strings.HasSuffix(player.Name, initPlayer.Name) {
					initEnhancedToonDescMap := enhancedToonDescMap[toonKey]

					if initEnhancedToonDescMap.Name == "" {
						initEnhancedToonDescMap.Name = initPlayer.Name
					}

					initEnhancedToonDescMap.HighestLeague = initPlayer.HighestLeague
					initEnhancedToonDescMap.IsInClan = initPlayer.IsInClan
					initEnhancedToonDescMap.ClanTag = initPlayer.ClanTag

					enhancedToonDescMap[toonKey] = initEnhancedToonDescMap
				}
			}
		}
	}

	return enhancedToonDescMap
}

// cleanMessageEvents copies the message events,
// has the capability of removing unescessary fields.
func cleanMessageEvents(replayData *rep.Rep) []s2prot.Struct {
	// Constructing a clean MessageEvents without unescessary fields:
	var messageEventsStructs []s2prot.Struct
	for _, messageEvent := range replayData.MessageEvts {
		messageEventsStructs = append(messageEventsStructs, messageEvent.Struct)
	}
	log.Info("Defined cleanMessageEvents struct")
	return messageEventsStructs
}

// cleanGameEvents copies the game events,
// has the capability of removing unescessary fields.
func cleanGameEvents(replayData *rep.Rep) []s2prot.Struct {
	// Constructing a clean GameEvents without unescessary fields:
	var gameEventsStructs []s2prot.Struct
	for _, gameEvent := range replayData.GameEvts {
		gameEventsStructs = append(gameEventsStructs, gameEvent.Struct)
	}
	log.Info("Defined cleanGameEvents struct")
	return gameEventsStructs
}

// cleanTrackerEvents copies the tracker events,
// has the capability of removing unescessary fields.
func cleanTrackerEvents(replayData *rep.Rep) []s2prot.Struct {
	// Constructing a clean TrackerEvents without unescessary fields:
	var trackerEventsStructs []s2prot.Struct
	for _, trackerEvent := range replayData.TrackerEvts.Evts {

		// https://github.com/Kaszanas/SC2InfoExtractorGo/issues/41
		if trackerEvent.Struct["evtTypeName"] == "PlayerStats" {

			// Get stats:
			stats := trackerEvent.Struct["stats"].(s2prot.Struct)

			// Get values:
			foodUsed := stats["scoreValueFoodUsed"].(int64) / 4096
			foodMade := stats["scoreValueFoodMade"].(int64) / 4096

			// Overwrite values:
			trackerEvent.Struct["stats"].(s2prot.Struct)["scoreValueFoodUsed"] = foodUsed
			trackerEvent.Struct["stats"].(s2prot.Struct)["scoreValueFoodMade"] = foodMade
		}

		trackerEventsStructs = append(trackerEventsStructs, trackerEvent.Struct)
	}
	log.Info("Defined cleanTrackerEvents struct")
	return trackerEventsStructs
}
