package sc2_map_processing

import (
	"os"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	log "github.com/sirupsen/logrus"
)

func CheckProcessed(
	files []string,
	processedFiles persistent_data.DownloadedMapsReplaysToFileInfo,
) ([]string, bool) {

	filteredFilesToProcess := []string{}
	alreadyProcessedFiles := processedFiles.DownloadedMapsForFiles

	for _, file := range files {

		// TODO: This logic should be moved before getting the list of all files:
		// Check if the replay was already processed:
		fileInfo, err := os.Stat(file)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"replayFile": file,
			}).Error("Failed to get file info.")
			return []string{}, false
		}

		// If the file was already processed, and was not modified since,
		// then it will be skipped from getting the map URL:
		fileInfoToCheck, alreadyProcessed := alreadyProcessedFiles[file]
		if alreadyProcessed {
			// Check if the file was modified since the last time it was processed:
			if persistent_data.CheckFileInfoEq(
				fileInfo,
				fileInfoToCheck.(persistent_data.FileInformationToCheck),
			) {
				// It is the same so continue
				log.WithField("file", file).
					Warning("This replay was already processed, map should be available, continuing!")
				continue
			}
			// It wasn't the same so the replay should be processed again:
			log.WithField("file", file).
				Warn("Replay was modified since the last time it was processed! Processing again.")

			filteredFilesToProcess = append(filteredFilesToProcess, file)

		}
	}

	return filteredFilesToProcess, true
}
