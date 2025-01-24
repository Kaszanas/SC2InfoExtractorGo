package downloader

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/chunk_utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

func MapDownloaderPipeline(
	cliFlags utils.CLIFlags,
	files []string,
	downloadedMapsForReplaysFilepath string,
	foreignToEnglishMappingFilepath string,
) map[string]string {

	// Create maps directory if it doesn't exist:
	err := file_utils.GetOrCreateDirectory(cliFlags.MapsDirectory)
	if err != nil {
		log.WithField("error", err).Error("Failed to create maps directory.")
		return nil
	}

	// REVIEW: Start Review:
	existingMapFilesSet, err := file_utils.ExistingFilesSet(
		cliFlags.MapsDirectory, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to get existing map files set.")
		return nil
	}

	// STAGE ONE PRE-PROCESS:
	// Get all map URLs into a set:
	URLToFileNameMap, downloadedMapsForReplays, err := sc2_map_processing.
		GetAllReplaysMapURLs(
			files,
			downloadedMapsForReplaysFilepath,
			cliFlags,
		)
	if err != nil {
		log.WithField("error", err).Error("Failed to get all map URLs.")
		return nil
	}

	// Check which of the files were previously processed and exclude them:
	filesWithoutMaps, ok := sc2_map_processing.CheckProcessed(
		files,
		downloadedMapsForReplays,
	)
	if !ok {
		log.Error("Failed to check which files were previously processed.")
		return nil
	}

	// TODO: Chunk the files without maps, these are the only ones that should be
	// processed with the downloader:
	filesWithoutMapsChunks, ok := chunk_utils.GetChunkListAndPackageBool(
		filesWithoutMaps,
		cliFlags.NumberOfPackages,
		cliFlags.NumberOfThreads,
		len(filesWithoutMaps),
	)
	if !ok {
		log.Error("Failed to get chunks for processing files for the map downloader.")
		return nil
	}

	// TODO: Verify how to create a new main function that will be a standalone
	// map and dependency downloader with specific exposed functions for the
	// sc2infoextractorgo.

	// STAGE-TWO PRE-PROCESS: Attempt downloading all SC2 maps from the read replays.
	// Download all SC2 maps from the replays if they were not processed before:

	// Shared state for the downloader:
	downloadedMapFilesSet := make(map[string]struct{})
	downloaderSharedState, err := NewDownloaderSharedState(
		cliFlags.MapsDirectory,
		existingMapFilesSet,
		downloadedMapFilesSet,
		cliFlags.NumberOfThreads*2)
	defer downloaderSharedState.WorkerPool.StopAndWait()
	if err != nil {
		log.WithField("error", err).Error("Failed to create downloader shared state.")
		return nil
	}

	existingMapFilesSet, err = DownloadAllSC2Maps(
		&downloaderSharedState,
		downloadedMapsForReplays,
		downloadedMapsForReplaysFilepath,
		URLToFileNameMap,
		filesWithoutMapsChunks,
		cliFlags,
	)
	if err != nil {
		log.WithField("error", err).Error("Failed to download all SC2 maps.")
		return nil
	}

	// STAGE-Three PRE-PROCESS:
	// Read all of the map names from the drive and create a mapping
	// from foreign to english names:
	progressBarReadLocalizedData := utils.NewProgressBar(
		len(existingMapFilesSet),
		"[3/4] Reading map names from drive: ",
	)
	mainForeignToEnglishMapping := make(map[string]string)
	for existingMapFilepath := range existingMapFilesSet {

		foreignToEnglishMapping, err := sc2_map_processing.
			ReadLocalizedDataFromMapGetForeignToEnglishMapping(
				existingMapFilepath,
				progressBarReadLocalizedData,
			)
		if err != nil {
			log.WithField("error", err).
				Error("Error reading map name from drive. Map could not be processed")
			return nil
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

	// REVIEW: Finish Review

}
