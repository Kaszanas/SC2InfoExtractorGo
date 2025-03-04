package downloader

import (
	"net/url"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

func MapDownloaderPipeline(
	files []string,
	foreignToEnglishMappingFilepath string,
	cliFlags utils.CLIFlags,
) map[string]string {

	if !cliFlags.SkipMapsDownload {
		// Create maps directory if it doesn't exist:
		err := file_utils.GetOrCreateDirectory(cliFlags.MapsDirectory)
		if err != nil {
			log.WithField("error", err).Error("Failed to create maps directory.")
			return nil
		}

		// REVIEW: Start Review:
		// STAGE ONE PRE-PROCESS:
		// Get all map URLs into a set:
		URLsToDownload, err := getURLsForMissingMaps(
			files,
			cliFlags,
		)
		if err != nil {
			log.WithField("error", err).Error("Failed to get URLs for missing maps.")
			return nil
		}

		// TODO: Verify how to create a new main function that will be a standalone
		// map and dependency downloader with specific exposed functions for the
		// sc2infoextractorgo.

		// STAGE-TWO PRE-PROCESS: Attempt downloading all SC2 maps from the read replays.
		// Download all SC2 maps from the replays if they were not processed before:
		downloadMissingMaps(URLsToDownload, cliFlags)
	}

	// STAGE-Three PRE-PROCESS:
	// Read all of the map names from the drive and create a mapping
	// from foreign to english names:
	mainForeignToEnglishMapping := readMapNamesFromMapFiles(
		foreignToEnglishMappingFilepath,
		cliFlags,
	)

	// REVIEW: Finish Review
	return mainForeignToEnglishMapping
}

func getURLsForMissingMaps(
	files []string,
	cliFlags utils.CLIFlags,
) (map[url.URL]string, error) {

	existingMapFilesSet, err := file_utils.ExistingFilesSet(
		cliFlags.MapsDirectory, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to get existing map files set.")
		return nil, err
	}

	URLsToDownload, err := sc2_map_processing.
		GetAllReplaysMapURLs(
			files,
			existingMapFilesSet,
			cliFlags,
		)
	if err != nil {
		log.WithField("error", err).Error("Failed to get all map URLs.")
		return nil, err
	}

	return URLsToDownload, nil
}

// Define a struct to represent the tuple
type URLToFileTuple struct {
	URL      url.URL
	Filename string
}

func downloadMissingMaps(
	URLsToDownload map[url.URL]string,
	cliFlags utils.CLIFlags,
) {

	// Shared state for the downloader:
	// existingMapFilesSet is required here to check if the map can be read.
	// In case of corrupted maps they will be redownloaded:
	downloaderSharedState, err := NewDownloaderSharedState(cliFlags)
	defer downloaderSharedState.WorkerPool.StopAndWait()
	if err != nil {
		log.WithField("error", err).Error("Failed to create downloader shared state.")
		return
	}

	err = DownloadAllSC2Maps(
		&downloaderSharedState,
		URLsToDownload,
		cliFlags,
	)
	if err != nil {
		log.WithField("error", err).Error("Failed to download all SC2 maps.")
		return
	}
}

func readMapNamesFromMapFiles(
	foreignToEnglishMappingFilepath string,
	cliFlags utils.CLIFlags,
) map[string]string {

	existingMapFilesSet, err := file_utils.ExistingFilesSet(
		cliFlags.MapsDirectory, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to get existing map files set.")
		return nil
	}

	progressBarReadLocalizedData := utils.NewProgressBar(
		len(existingMapFilesSet),
		"[3/4] Reading english map names from map files: ",
	)
	mainForeignToEnglishMapping := make(map[string]string)
	for existingMapFilepath := range existingMapFilesSet {

		foreignToEnglishMapping, err := sc2_map_processing.
			ReadLocalizedDataFromMapGetForeignToEnglishMapping(
				existingMapFilepath,
				progressBarReadLocalizedData,
			)
		if err != nil {
			log.WithFields(log.Fields{
				"error":               err,
				"existingMapFilepath": existingMapFilepath,
			}).Error("Error reading map name from drive. Map could not be processed")
			continue
		}

		// Fill out the mapping, these maps won't be opened again:
		for foreignName, englishName := range foreignToEnglishMapping {
			mainForeignToEnglishMapping[foreignName] = englishName
		}
	}
	// Save the mapping to the drive:
	err = sc2_map_processing.SaveForeignToEnglishMappingToDrive(
		foreignToEnglishMappingFilepath,
		mainForeignToEnglishMapping,
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to save foreign to english mapping to drive.")
		return nil
	}
	return mainForeignToEnglishMapping
}
