package file_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ReadOrCreateFile receives a filepath, creates a file if it doesn't exist.
// Otherwise it reads the file and returns the file and its contents.
func ReadOrCreateFile(filepath string) (os.File, []byte, error) {

	log.WithField("filepath", filepath).Info("Entered ReadOrCreateFile()")

	_, err := os.Stat(filepath)
	// File Doesn't exist:
	if os.IsNotExist(err) {
		// Create the file:
		createdFile, err := CreateTruncateFile(filepath)
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err,
				"filepath": filepath,
			}).Fatal("Failed to create the file!")
			return os.File{}, nil, err
		}
		return createdFile, []byte{}, nil
	}

	// Otherwise file exists:
	file, err := os.OpenFile(filepath, os.O_RDWR, 0666)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to open the file!")
		return os.File{}, nil, err
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to read the file!")
		return os.File{}, nil, err
	}

	log.Debug("Finished ReadOrCreateFile()")
	return *file, contents, nil
}

// CreateTruncateFile receives a filepath and creates a file if it doesn't exist.
func CreateTruncateFile(filePath string) (os.File, error) {

	log.Debug("Entered CreateTruncateFile()")

	createdOrReadFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filePath": filePath,
		}).Fatal("Failed to create or open the file!")
		return os.File{}, err
	}

	log.Debug("Finished CreateTruncateFile()")
	return *createdOrReadFile, nil
}

// GetOrCreateDirectory receives the path to the
// maps directory and creates it if it doesn't exist.
func GetOrCreateDirectory(pathToMapsDirectory string) error {

	log.Debug("Entered CreateMapsDirectory()")

	// Create the maps directory:
	err := os.MkdirAll(pathToMapsDirectory, 0777)
	if os.IsExist(err) {
		log.Info("The maps directory already exists!")
		return nil
	}
	if err != nil {
		log.WithField("error", err).
			Fatal("failed to create the maps directory!")
		return fmt.Errorf("failed to create the maps directory: %v", err)
	}

	log.Debug("Finished GetOrCreateMapsDirectory()")
	return nil
}

// UnmarshalJSONMapping wraps around unmarshalLocaleFile and returns
// an empty map[string]interface{} if it fails to unmarshal the original locale mapping file.
func UnmarshalJSONMapping(
	pathToMappingFile string,
) (map[string]interface{}, error) {
	log.Debug("Entered unmarshalLocaleMapping()")

	unmarshalledMap := make(map[string]interface{})
	err := UnmarshalJsonFile(pathToMappingFile, &unmarshalledMap)
	if err != nil {
		log.WithField("pathToMappingFile", pathToMappingFile).
			Error("Failed to open and unmarshal the mapping file!")
		return unmarshalledMap, err
	}

	log.Debug("Finished unmarshalLocaleMapping()")
	return unmarshalledMap, nil
}

// UnmarshalJsonFile deals with every possible opening and unmarshalling
// problem that might occur when unmarshalling a localization file
// supplied by: https://github.com/Kaszanas/SC2MapLocaleExtractor
func UnmarshalJsonFile(
	filepath string,
	mapToPopulate *map[string]interface{},
) error {
	log.Debug("Entered unmarshalJsonFile()")

	var file, err = os.Open(filepath)
	if err != nil {
		log.WithField("fileError", err.Error()).
			Info("Failed to open the JSON file.")
		return err
	}
	defer file.Close()

	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		log.WithField("readError", err.Error()).
			Info("Failed to read the JSON file.")
		return err
	}

	err = json.Unmarshal([]byte(jsonBytes), &mapToPopulate)
	if err != nil {
		log.WithField("jsonMarshalError", err.Error()).
			Info("Could not unmarshal the JSON file.")
	}

	log.Debug("Finished unmarshalJsonFile()")
	return nil
}

// SaveReplayJSONFileToDrive is a helper function that takes
// the json string of a StarCraft II replay and writes it to drive.
func SaveReplayJSONFileToDrive(
	replayString string,
	replayFile string,
	absolutePathOutputDirectory string) bool {

	_, replayFileNameWithExt := filepath.Split(replayFile)

	replayFileName := strings.TrimSuffix(
		replayFileNameWithExt,
		filepath.Ext(replayFileNameWithExt),
	)

	jsonAbsPath := filepath.Join(absolutePathOutputDirectory, replayFileName+".json")
	jsonBytes := []byte(replayString)

	err := os.WriteFile(jsonAbsPath, jsonBytes, 0777)
	if err != nil {
		log.WithField("replayFile", replayFile).
			Error("Failed to write .json to drive!")
		return false
	}

	return true
}
