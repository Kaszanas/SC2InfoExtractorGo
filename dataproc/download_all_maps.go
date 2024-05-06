package dataproc

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

// downloadAllSC2Maps download all of the maps from the replays
// if the replays were not processed before.
func donwloadAllSC2MapsHandleProcessedReplays(
	downloaderSharedState *downloader.DownloaderSharedState,
	mapsDirectory string,
	processedReplaysFilepath string,
	fileChunks [][]string,
	cliFlags utils.CLIFlags,
) (map[string]struct{}, error) {

	log.WithFields(log.Fields{
		"mapsDirectory":            mapsDirectory,
		"processedReplaysFilepath": processedReplaysFilepath}).
		Info("Entered downloadAllSC2Maps()")

	defer downloaderSharedState.WorkerPool.StopAndWait()

	processedReplays, err := persistent_data.OpenOrCreateProcessedReplaysToFileInfo(
		processedReplaysFilepath,
		mapsDirectory,
		fileChunks,
	)
	if err != nil {
		return nil, err
	}

	for _, chunk := range fileChunks {
		for _, replayFile := range chunk {
			// Check if the replay was already processed:
			// This assumes availability of the map in the maps directory:
			_, ok := processedReplays.CheckIfReplayWasProcessed(replayFile)
			if ok {
				log.Debug("Replay was already processed, continuing.")
				continue
			}
			// If it wasn't, open the replay, get map information,
			// download the map, and save it to the drive.
			err := processedReplays.DownloadMapAddReplayToProcessed(
				replayFile,
				downloaderSharedState,
			)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"replayFile": replayFile,
				}).Error("Failed to download the map. May have to be downloaded manually!")
				continue
			}
		}
	}

	// Wait Stop and wait without defer,
	// all of the maps need to finish downloading before the processing starts:
	downloaderSharedState.WorkerPool.StopAndWait()

	// Save the processed replays to the file:
	err = processedReplays.SaveProcessedReplaysFile(processedReplaysFilepath)
	if err != nil {
		log.WithField("processedReplaysFile", processedReplaysFilepath).
			Error("Failed to save the processed replays file.")
		return nil, err
	}

	// Get the list of maps after the download finishes:
	existingMapFilesSet, err := file_utils.ExistingFilesSet(mapsDirectory, ".s2ma")
	if err != nil {
		log.WithField("mapsDirectory", mapsDirectory).
			Error("Failed to get the set of existing map files.")
		return nil, err
	}

	// TODO: Remove files that were already processed from the existingMapFilesSet
	// TODO: using information from persistent_data.ProcessedReplaysToMaps

	return existingMapFilesSet, nil
}
