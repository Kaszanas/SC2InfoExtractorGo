package persistent_data

import (
	"fmt"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// ProcessedReplaysToMaps
type ProcessedReplaysToMaps struct {
	ProcessedFiles map[string]interface{} `json:"processedReplays"`
}

func (prtm *ProcessedReplaysToMaps) CheckIfReplayWasProcessed(replayPath string) bool {
	_, ok := prtm.ProcessedFiles[replayPath]
	return ok
}

func (prtm *ProcessedReplaysToMaps) AddReplayToProcessed(
	replayPath string,
	downloaderSharedState *downloader.DownloaderSharedState,
) error {

	replayData, err := rep.NewFromFile(replayPath)
	if err != nil {
		log.WithFields(log.Fields{
			"err":        err,
			"replayPath": replayPath}).
			Error("Failed to get replay data.")
		return err
	}

	// Getting map URL and hash before mutexing, this operation is not thread safe:
	mapURL, mapHashAndExtension, ok := sc2_map_processing.
		GetMapURLAndHashFromReplayData(replayData)
	if !ok {
		log.Error("getMapURLAndHashFromReplayData() failed.")
		return fmt.Errorf("getMapURLAndHashFromReplayData() failed")
	}

	ok = downloader.GetEnglishMapNameDownloadIfNotExists(
		downloaderSharedState,
		mapHashAndExtension,
		mapURL)
	if !ok {
		log.WithField("file", replayPath).
			Error("Failed to get English map name.")

		log.Error("getEnglishMapNameDownloadIfNotExists() failed.")
		return fmt.Errorf("getEnglishMapNameDownloadIfNotExists() failed")
	}

	return nil
}

// NewProcessedReplaysToMaps returns empty ProcessingInfo struct.
func NewProcessedReplaysToMaps(
	filepath string,
	mapDirectory string,
) (ProcessedReplaysToMaps, error) {

	// check if the file exists:
	mapToPopulate := make(map[string]interface{})
	// TODO: These _ skipped variables could be used in unmarshalling:
	_, _, err := file_utils.ReadOrCreateFile(filepath)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to read or create the processed_replays.json file.")
		return ProcessedReplaysToMaps{}, err
	}
	err = file_utils.UnmarshalJsonFile(filepath, &mapToPopulate)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to unmarshal the processed_replays.json file.")
		return ProcessedReplaysToMaps{}, err
	}

	// Check if number of unique values in the map is equal to the number of files in the map directory
	// if not then delete and start from scratch.
	uniqueValues := make(map[interface{}]struct{})
	for key := range mapToPopulate {
		uniqueValues[key] = struct{}{}
	}

	// list the files in the map directory:
	files := file_utils.ListFiles(mapDirectory, ".s2ma")

	//
	if len(uniqueValues) != len(files) {
		log.WithFields(log.Fields{
			"lenUniqueValues": len(uniqueValues),
			"lenFiles":        len(files),
		}).
			Info("Unique values in the map are not equal to the number of files in the map directory.")
		_, err := file_utils.CreateTruncateFile(filepath)
		if err != nil {
			log.WithField("error", err).
				Error("Failed to create the processed_replays.json file.")
			return ProcessedReplaysToMaps{}, err
		}
	}

	return ProcessedReplaysToMaps{
		ProcessedFiles: mapToPopulate,
	}, nil
}
