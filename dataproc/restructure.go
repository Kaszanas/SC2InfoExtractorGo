package dataproc

import (
	"fmt"
	"strings"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/cleanup"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// redefineReplayStructure moves arbitrary data into different data structures.
func redifineReplayStructure(
	replayData *rep.Rep,
	englishToForeignMapping map[string]string,
) (replay_data.CleanedReplay, bool) {

	log.Info("Entered redefineReplayStructure()")

	cleanHeader := cleanup.CleanHeader(replayData)
	cleanGameDescription, ok := cleanup.CleanGameDescription(replayData)
	if !ok {
		return replay_data.CleanedReplay{}, false
	}
	cleanInitData, cleanedUserInitDataList, ok := cleanup.CleanInitData(
		replayData,
		cleanGameDescription)
	if !ok {
		return replay_data.CleanedReplay{}, false
	}
	cleanDetails, detailsReplayMapField := cleanup.CleanDetails(replayData)
	cleanMetadata, metadataReplayMapField := cleanup.CleanMetadata(replayData)

	mapFields := []replay_data.ReplayMapField{
		detailsReplayMapField,
		metadataReplayMapField,
	}
	ok = adjustMapName(mapFields, englishToForeignMapping, &cleanMetadata)
	if !ok {
		log.Error("Failed to adjust map name!")
		return replay_data.CleanedReplay{}, false
	}

	// This is used for older replay versions where some fields are missing:
	ok = adjustGameVersion(&cleanHeader, &cleanMetadata)
	if !ok {
		log.Error("Failed to adjust game version!")
		return replay_data.CleanedReplay{}, false
	}

	enhancedToonDescMap, ok := cleanup.CleanToonDescMap(replayData, cleanedUserInitDataList)
	if !ok {
		log.Error("Failed to clean toon desc map!")
		return replay_data.CleanedReplay{}, false
	}

	messageEventsStructs := cleanup.CleanMessageEvents(replayData)
	gameEventsStructs := cleanup.CleanGameEvents(replayData)
	trackerEventsStructs := cleanup.CleanTrackerEvents(replayData)

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

// getVersionElements splits the version string into its elements,
// returns an error if the version string does not contain 4 elements.
// Otherwise returns a slice of strings corresponding to the version elements.
func getVersionElements(version string) ([]string, error) {

	versionElements := strings.Split(version, ".")
	if len(versionElements) != 4 {
		return nil, fmt.Errorf("version string does not contain 4 elements")
	}

	return versionElements, nil
}

// adjustGameVersion adjusts the game version if it is missing in either header or metadata.
func adjustGameVersion(
	cleanHeader *replay_data.CleanedHeader,
	cleanMetadata *replay_data.CleanedMetadata,
) bool {

	headerVersionOk := cleanHeader.Version != ""
	metadataVersionOk := cleanMetadata.GameVersion != ""

	if !headerVersionOk && !metadataVersionOk {
		log.Error("Both game version fields are empty!")
		return false
	}

	if !headerVersionOk && metadataVersionOk {
		log.Info("Found empty game version in metadata, fill out header!")
		// Header version exists, fill out metadata:

		versionElements, err := getVersionElements(cleanMetadata.GameVersion)
		if err != nil {
			log.WithField("error", err.Error()).
				Error("Failed to split version string into elements from header!")
			return false
		}

		// Base build is the last element of the version string:
		cleanMetadata.BaseBuild = fmt.Sprintf("Base%s", versionElements[len(versionElements)-1])
		cleanMetadata.DataBuild = versionElements[len(versionElements)-1]
		cleanMetadata.GameVersion = cleanHeader.Version

		return true
	}

	if headerVersionOk && !metadataVersionOk {
		log.Info("Found empty game version in header, fill out metadata!")
		// Metadata version exists, fill out header:
		// If the game version is not available in metadata,
		// we will use the one from header:

		versionElements, err := getVersionElements(cleanHeader.Version)
		log.Info("versionElements: ", versionElements)
		if err != nil {
			log.WithField("error", err.Error()).
				Error("Failed to split version string into elements from metadata!")
			return false
		}

		cleanMetadata.GameVersion = cleanHeader.Version
		cleanMetadata.BaseBuild = fmt.Sprintf("Base%s", versionElements[len(versionElements)-1])
		cleanMetadata.DataBuild = versionElements[len(versionElements)-1]

		return true
	}

	log.Info("Both metadata and header contains game version")
	return true
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
