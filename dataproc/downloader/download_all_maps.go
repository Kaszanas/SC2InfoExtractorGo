package downloader

import (
	"net/url"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

// DownloadAllSC2Dependencies download all of the dependencies from the replays
// if the replays were not processed before.
func DownloadAllSC2Dependencies(
	downloaderSharedState *DownloaderSharedState,
	URLsToDownload map[url.URL]sc2_map_processing.ReplayFilenameIsMap,
	cliFlags utils.CLIFlags,
) error {

	log.WithFields(log.Fields{
		"dependencyDirectory": cliFlags.DependencyDirectory},
	).Debug("Entered DownloadAllSC2Dependencies()")

	defer downloaderSharedState.WorkerPool.StopAndWait()

	// Progress bar:
	progressBarDownloadDependencies := utils.NewProgressBar(
		len(URLsToDownload),
		"[2/4] Downloading dependencies: ",
	)
	defer progressBarDownloadDependencies.Close()
	for url, filenameAndIsMap := range URLsToDownload {

		// If the replay was not processed previosly,
		// open the replay, get map information,
		// download the map, and save it to the drive.
		err := DownloadDependencyIfNotExists(
			downloaderSharedState,
			filenameAndIsMap,
			url,
			progressBarDownloadDependencies,
		)
		if err != nil {
			log.WithFields(log.Fields{
				"mapURL":                     url.String(),
				"dependencyHashAndExtension": filenameAndIsMap.DependencyFilename,
			}).Error("Failed to download the map.")
		}
	}
	// Wait Stop and wait without defer,
	// all of the dependencies need to finish downloading before the processing starts:
	downloaderSharedState.WorkerPool.StopAndWait()
	progressBarDownloadDependencies.Close()

	log.Debug("Finished DownloadAllSC2Dependencies()")
	return nil
}
