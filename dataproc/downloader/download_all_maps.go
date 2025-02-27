package downloader

import (
	"net/url"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

// downloadAllSC2Maps download all of the maps from the replays
// if the replays were not processed before.
func DownloadAllSC2Maps(
	downloaderSharedState *DownloaderSharedState,
	URLsToDownload map[url.URL]string,
	cliFlags utils.CLIFlags,
) error {

	log.WithFields(log.Fields{
		"mapsDirectory": cliFlags.MapsDirectory},
	).Debug("Entered downloadAllSC2Maps()")

	defer downloaderSharedState.WorkerPool.StopAndWait()

	// Progress bar:
	progressBarDownloadMaps := utils.NewProgressBar(
		len(URLsToDownload),
		"[2/4] Downloading maps: ",
	)
	defer progressBarDownloadMaps.Close()
	for url, mapHashAndExtension := range URLsToDownload {

		// If the replay was not processed previosly,
		// open the replay, get map information,
		// download the map, and save it to the drive.
		err := DownloadMapIfNotExists(
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

	log.Debug("Finished downloadAllSC2Maps()")
	return nil
}
