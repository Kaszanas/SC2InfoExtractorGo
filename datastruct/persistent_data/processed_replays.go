package persistent_data

import (
	"encoding/json"
	"io/fs"
	"os"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
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
				if CheckFileInfoEq(fileInfo, fileInfoToCheck) {
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

// CheckFileInfoEq compares the fs.FileInfo contents with FileInfoToCheck.
// Returns true if the contents are the same.
// This is used to verify if the file changed since last processing.
func CheckFileInfoEq(
	fileInfo fs.FileInfo,
	fileInfoToCheck FileInformationToCheck,
) bool {
	return fileInfo.ModTime().Unix() == fileInfoToCheck.LastModified &&
		fileInfo.Size() == fileInfoToCheck.Size
}

// ConvertToSyncMap Converts a ProcessedReplaysToFileInfo to a sync.Map.
func (prtm *ProcessedReplaysToFileInfo) ConvertToSyncMap() *sync.Map {
	syncMap := &sync.Map{}
	for key, value := range prtm.ProcessedFiles {
		syncMap.Store(key, value)
	}
	return syncMap
}

// FromSyncMapToProcessedReplaysToFileInfo Converts a
// sync.Map to a ProcessedReplaysToFileInfo.
func FromSyncMapToProcessedReplaysToFileInfo(
	syncMap *sync.Map,
) ProcessedReplaysToFileInfo {
	processedReplays := make(map[string]interface{})
	syncMap.Range(func(key, value interface{}) bool {
		processedReplays[key.(string)] = value.(FileInformationToCheck)
		return true
	})

	return ProcessedReplaysToFileInfo{
		ProcessedFiles: processedReplays,
	}
}

// CheckIfReplayWasProcessed checks if the replay was processed before.
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

// AddReplayToProcessed adds a replay with its file information to the processed replays.
// used to check if the replay was processed before.
func (prtm *ProcessedReplaysToFileInfo) AddReplayToProcessed(
	replayPath string,
	fileInfo fs.FileInfo,
) {
	prtm.ProcessedFiles[replayPath] = FileInformationToCheck{
		LastModified: fileInfo.ModTime().Unix(),
		Size:         fileInfo.Size(),
	}
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
