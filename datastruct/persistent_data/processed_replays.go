package persistent_data

import (
	"encoding/json"
	"io/fs"
	"os"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

// DownloadedMapsReplaysToFileInfo Used to store replays that were processed in previous runs.
// Holds file information in the values of the map to compare against current files.
// This is so that the maps for them are not re-downloaded.
// This is used in a pre-processing step,
// before the final processing of the replays and data extraction is performed.
type DownloadedMapsReplaysToFileInfo struct {
	DownloadedMapsForFiles map[string]any `json:"downloadedMapsForFiles"`
}

type FileInformationToCheck struct {
	LastModified int64 `json:"lastModified"`
	Size         int64 `json:"size"`
}

// OpenOrCreateDownloadedMapsForReplaysToFileInfo Initializes and populates
// a DownloadedMapsReplaysToFileInfo structure.
func OpenOrCreateDownloadedMapsForReplaysToFileInfo(
	filepath string,
	mapsDirectory string,
	files []string,
) (DownloadedMapsReplaysToFileInfo, []string, error) {

	replaysToProcess := []string{}
	// check if the file exists:
	mapToPopulateFromPersistentJSON := make(map[string]any)
	_, _, err := file_utils.ReadOrCreateFile(filepath)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to read or create the processed_replays.json file.")
		return DownloadedMapsReplaysToFileInfo{}, replaysToProcess, err
	}
	err = file_utils.UnmarshalJsonFile(filepath, &mapToPopulateFromPersistentJSON)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to unmarshal the processed_replays.json file.")
		return DownloadedMapsReplaysToFileInfo{}, replaysToProcess, err
	}

	alreadyProcessed := DownloadedMapsReplaysToFileInfo{
		DownloadedMapsForFiles: mapToPopulateFromPersistentJSON,
	}

	for _, replayFile := range files {
		fileInfo, err := os.Stat(replayFile)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"replayFile": replayFile,
			}).Error("Failed to get file info.")
			return DownloadedMapsReplaysToFileInfo{}, replaysToProcess, err
		}
		fileInfoToCheck, ok := alreadyProcessed.CheckIfReplayWasProcessed(replayFile)
		// Replay was processed so no need to check it again:
		if ok {
			// Check if the file was modified since the last time it was processed:
			if CheckFileInfoEq(fileInfo, fileInfoToCheck) {
				// It is the same so continue
				continue
			}
		}
		// It wasn't the same so the replay should be processed again:
		replaysToProcess = append(replaysToProcess, replayFile)
		delete(alreadyProcessed.DownloadedMapsForFiles, replayFile)
	}

	return alreadyProcessed, replaysToProcess, nil
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

// ConvertToSyncMap Converts a DownloadedMapsReplaysToFileInfo to a sync.Map.
func (prtm *DownloadedMapsReplaysToFileInfo) ConvertToSyncMap() *sync.Map {
	syncMap := &sync.Map{}
	for key, value := range prtm.DownloadedMapsForFiles {
		valueMap := value.(map[string]any)
		// JSON is unmarshaled to float64 so we need to cast it to int64:
		infoToCheck := FileInformationToCheck{
			LastModified: int64(valueMap["lastModified"].(float64)),
			Size:         int64(valueMap["size"].(float64)),
		}
		syncMap.Store(key, infoToCheck)
	}
	return syncMap
}

// FromSyncMapToDownloadedMapsForReplaysToFileInfo Converts a
// sync.Map to a DownloadedMapsReplaysToFileInfo.
func FromSyncMapToDownloadedMapsForReplaysToFileInfo(
	syncMap *sync.Map,
) DownloadedMapsReplaysToFileInfo {
	downloadedMapsForReplays := make(map[string]any)
	syncMap.Range(func(key, value interface{}) bool {

		keyStr := key.(string)
		valueFileInformationToCheck, _ := value.(FileInformationToCheck)

		downloadedMapsForReplays[keyStr] = valueFileInformationToCheck
		return true
	})

	return DownloadedMapsReplaysToFileInfo{
		DownloadedMapsForFiles: downloadedMapsForReplays,
	}
}

// CheckIfReplayWasProcessed checks if the replay was processed before.
func (prtm *DownloadedMapsReplaysToFileInfo) CheckIfReplayWasProcessed(
	replayPath string,
) (FileInformationToCheck, bool) {
	fileInfo, ok := prtm.DownloadedMapsForFiles[replayPath]
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
func (prtm *DownloadedMapsReplaysToFileInfo) AddReplayToProcessed(
	replayPath string,
	fileInfo fs.FileInfo,
) {
	prtm.DownloadedMapsForFiles[replayPath] = FileInformationToCheck{
		LastModified: fileInfo.ModTime().Unix(),
		Size:         fileInfo.Size(),
	}
}

func (prtm *DownloadedMapsReplaysToFileInfo) SaveDownloadedMapsForReplaysFile(
	filepath string,
) error {

	jsonBytes, err := json.Marshal(prtm.DownloadedMapsForFiles)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to marshal the downloadedMapsForReplays map.")
		return err
	}

	downloadedMapsForReplaysFile, err := file_utils.CreateTruncateFile(filepath)
	if err != nil {
		log.Error("Failed to create the package summary file!")
		return err
	}

	_, err = downloadedMapsForReplaysFile.Write(jsonBytes)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to save the downloadedMapsForReplaysFile")
		return err
	}

	return nil
}
