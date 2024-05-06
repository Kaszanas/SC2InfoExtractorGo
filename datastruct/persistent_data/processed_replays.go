package persistent_data

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// ProcessedReplaysToFileInfo Used to store replays that were processed in previous runs.
// Holds file information in the values of the map to compare against current files.
// This is so that the maps for them are not re-downloaded.
// This is used in a pre-processing step,
// before the final processing of the replays and data extraction is performed.
type ProcessedReplaysToFileInfo struct {
	ProcessedFiles map[string]interface{} `json:"processedReplays"`
}

type FileInformationToCheck struct {
	LastModified int64 `json:"lastModified"`
	Size         int64 `json:"size"`
}

// OpenOrCreateProcessedReplaysToFileInfo Initializes and populates
// a ProcessedReplaysToFileInfo structure.
func OpenOrCreateProcessedReplaysToFileInfo(
	filepath string,
	mapDirectory string,
	fileChunks [][]string,
) (ProcessedReplaysToFileInfo, error) {

	// check if the file exists:
	mapToPopulateFromPersistentJSON := make(map[string]interface{})
	_, _, err := file_utils.ReadOrCreateFile(filepath)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to read or create the processed_replays.json file.")
		return ProcessedReplaysToFileInfo{}, err
	}
	err = file_utils.UnmarshalJsonFile(filepath, &mapToPopulateFromPersistentJSON)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to unmarshal the processed_replays.json file.")
		return ProcessedReplaysToFileInfo{}, err
	}

	prtm := ProcessedReplaysToFileInfo{
		ProcessedFiles: mapToPopulateFromPersistentJSON,
	}

	for _, chunk := range fileChunks {
		for _, replayFile := range chunk {
			fileInfo, err := os.Stat(replayFile)
			if err != nil {
				log.WithFields(log.Fields{
					"error":      err,
					"replayFile": replayFile,
				}).Error("Failed to get file info.")
				return ProcessedReplaysToFileInfo{}, err
			}
			fileInfoToCheck, ok := prtm.CheckIfReplayWasProcessed(replayFile)
			// Replay was processed so no need to check it again:
			if ok {
				// Check if the file was modified since the last time it was processed:
				if checkFileInfoEq(fileInfo, fileInfoToCheck) {
					// It is the same so continue
					continue
				}
			}
			// It wasn't the same so the replay should be processed again:
			delete(prtm.ProcessedFiles, replayFile)
		}
	}

	return prtm, nil
}

func checkFileInfoEq(
	fileInfo fs.FileInfo,
	fileInfoToCheck FileInformationToCheck,
) bool {
	return fileInfo.ModTime().Unix() == fileInfoToCheck.LastModified &&
		fileInfo.Size() == fileInfoToCheck.Size
}

func (prtm *ProcessedReplaysToFileInfo) CheckIfReplayWasProcessed(
	replayPath string,
) (FileInformationToCheck, bool) {
	fileInfo, ok := prtm.ProcessedFiles[replayPath]
	if !ok {
		return FileInformationToCheck{}, ok
	}

	fileInfoToCheck, ok := fileInfo.(FileInformationToCheck)
	if !ok {
		log.WithField("fileInfo", fileInfo).
			Error("Failed to cast fileInfo to FileInformationToCheck.")
		return FileInformationToCheck{}, ok
	}

	return fileInfoToCheck, ok
}

func (prtm *ProcessedReplaysToFileInfo) AddReplayToProcessed(
	replayPath string,
	fileInfo fs.FileInfo,
) {
	prtm.ProcessedFiles[replayPath] = FileInformationToCheck{
		LastModified: fileInfo.ModTime().Unix(),
		Size:         fileInfo.Size(),
	}
}

func (prtm *ProcessedReplaysToFileInfo) DownloadMapAddReplayToProcessed(
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

	err = downloader.DownloadMapIfNotExists(
		downloaderSharedState,
		mapHashAndExtension,
		mapURL)
	if err != nil {
		log.WithField("file", replayPath).
			Error("Failed to get English map name.")
		return fmt.Errorf("getEnglishMapNameDownloadIfNotExists() failed: %v", err)
	}

	fileInfo, err := os.Stat(replayPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"replayPath": replayPath,
		}).Error("Failed to get file info.")
		return err
	}

	// Map was downloaded successfully,
	// add the replay to the processed replays:
	fileInfoToCheck := FileInformationToCheck{
		LastModified: fileInfo.ModTime().Unix(),
		Size:         fileInfo.Size(),
	}
	prtm.ProcessedFiles[replayPath] = fileInfoToCheck

	return nil
}

func (prtm *ProcessedReplaysToFileInfo) SaveProcessedReplaysFile(
	filepath string,
) error {

	jsonBytes, err := json.Marshal(prtm.ProcessedFiles)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to marshal the processedReplays map.")
		return err
	}

	processedReplaysFile, err := file_utils.CreateTruncateFile(filepath)
	if err != nil {
		log.Error("Failed to create the package summary file!")
		return err
	}

	_, err = processedReplaysFile.Write(jsonBytes)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to save the processedReplaysFile")
		return err
	}

	return nil
}
