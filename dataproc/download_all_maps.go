package dataproc

import (
	"fmt"
	"net/url"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

// downloadAllSC2Maps download all of the maps from the replays
// if the replays were not processed before.
func DownloadAllSC2Maps(
	downloaderSharedState *downloader.DownloaderSharedState,
	mapsDirectory string,
	processedReplays persistent_data.ProcessedReplaysToFileInfo,
	allMapURLs map[url.URL]string,
	fileChunks [][]string,
	cliFlags utils.CLIFlags,
) (map[string]struct{}, error) {

	log.WithFields(log.Fields{
		"mapsDirectory":      mapsDirectory,
		"n_processedReplays": len(processedReplays.ProcessedFiles)}).
		Info("Entered downloadAllSC2Maps()")

	defer downloaderSharedState.WorkerPool.StopAndWait()

	for url, mapHashAndExtension := range allMapURLs {

		// If it wasn't, open the replay, get map information,
		// download the map, and save it to the drive.
		err := downloader.DownloadMapIfNotExists(
			downloaderSharedState,
			mapHashAndExtension,
			url,
		)
		if err != nil {
			log.WithFields(log.Fields{
				"mapURL":              url.String(),
				"mapHashAndExtension": mapHashAndExtension,
			}).Error("Failed to download the map.")
			return nil,
				fmt.Errorf("getEnglishMapNameDownloadIfNotExists() failed: %v", err)
		}
	}

	// Wait Stop and wait without defer,
	// all of the maps need to finish downloading before the processing starts:
	downloaderSharedState.WorkerPool.StopAndWait()

	// Get the list of maps after the download finishes:
	existingMapFilesSet, err := file_utils.ExistingFilesSet(
		mapsDirectory,
		".s2ma",
	)
	if err != nil {
		log.WithField("mapsDirectory", mapsDirectory).
			Error("Failed to get the set of existing map files.")
		return nil, err
	}

	return existingMapFilesSet, nil
}
