package dataproc

import (
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
	downloadedMapsForReplays persistent_data.DownloadedMapsReplaysToFileInfo,
	downloadedMapsForReplaysFilepath string,
	allMapURLs map[url.URL]string,
	fileChunks [][]string,
	cliFlags utils.CLIFlags,
) (map[string]struct{}, error) {

	log.WithFields(log.Fields{
		"mapsDirectory":              cliFlags.MapsDirectory,
		"n_downloadedMapsForReplays": len(downloadedMapsForReplays.DownloadedMapsForFiles)}).
		Info("Entered downloadAllSC2Maps()")

	defer downloaderSharedState.WorkerPool.StopAndWait()

	// Progress bar:
	progressBarDownloadMaps := utils.NewProgressBar(
		len(allMapURLs),
		"[2/4] Downloading maps: ",
	)
	defer progressBarDownloadMaps.Close()
	for url, mapHashAndExtension := range allMapURLs {

		// If it wasn't, open the replay, get map information,
		// download the map, and save it to the drive.
		err := downloader.DownloadMapIfNotExists(
			downloaderSharedState,
			mapHashAndExtension,
			url,
			progressBarDownloadMaps,
		)
		if err != nil {
			log.WithFields(log.Fields{
				"mapURL":              url.String(),
				"mapHashAndExtension": mapHashAndExtension,
			}).Error("Failed to download the map.")
		}
	}
	// Wait Stop and wait without defer,
	// all of the maps need to finish downloading before the processing starts:
	downloaderSharedState.WorkerPool.StopAndWait()
	progressBarDownloadMaps.Close()

	// Save the processed replays to the file:
	err := downloadedMapsForReplays.SaveDownloadedMapsForReplaysFile(
		downloadedMapsForReplaysFilepath,
	)
	if err != nil {
		log.WithField("downloadedMapsForReplaysFile",
			downloadedMapsForReplaysFilepath,
		).
			Error("Failed to save the processed replays file.")
		return nil, err
	}

	// Get the list of maps after the download finishes:
	existingMapFilesSet, err := file_utils.ExistingFilesSet(
		cliFlags.MapsDirectory,
		".s2ma",
	)
	if err != nil {
		log.WithField("mapsDirectory", cliFlags.MapsDirectory).
			Error("Failed to get the set of existing map files.")
		return nil, err
	}

	return existingMapFilesSet, nil
}
