package dataproc

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
)

// downloadAllSC2Maps download all of the maps from the replays
// if the replays were not processed before.
func donwloadAllSC2Maps(
	downloaderSharedState *downloader.DownloaderSharedState,
	mapsDirectory string,
	processedReplaysFile string,
	fileChunks [][]string,
	cliFlags utils.CLIFlags,
) ([]string, error) {

	defer downloaderSharedState.WorkerPool.StopAndWait()

	// TODO: Check if the replay was already touched by the previous runs of the program.
	// If it was, continue
	processedReplays, err := persistent_data.NewProcessedReplaysToMaps(
		processedReplaysFile,
		mapsDirectory,
	)
	if err != nil {
		return nil, err
	}

	for _, chunk := range fileChunks {
		for _, replayFile := range chunk {
			// Check if the replay was already processed:
			// This assumes availability of the map in the maps directory:
			if processedReplays.CheckIfReplayWasProcessed(replayFile) {
				continue
			}
			// If it wasn't, open the replay, get map information,
			// download the map, and save it to the drive.
			processedReplays.AddReplayToProcessed(replayFile, downloaderSharedState)
		}
	}

	// Wait Stop and wait without defer,
	// all of the maps need to finish downloading before the processing starts:
	downloaderSharedState.WorkerPool.StopAndWait()

	// Get the list of maps after the download finishes:
	existingMapFiles := file_utils.ListFiles(mapsDirectory, ".s2ma")

	return existingMapFiles, nil
}
