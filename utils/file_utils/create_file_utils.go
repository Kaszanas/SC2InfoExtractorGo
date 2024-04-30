package file_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	log "github.com/sirupsen/logrus"
)

// ReadOrCreateFile receives a filepath and creates a file if it doesn't exist.
func ReadOrCreateFile(filePath string, flag int) (os.File, []byte, error) {

	log.Info("Entered readOrCreateFile()")

	createdOrReadFile, err := os.OpenFile(filePath, flag, 0666)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filePath": filePath,
		}).Fatal("Failed to create or open the file!")
		return os.File{}, nil, err
	}
	byteValue, err := io.ReadAll(createdOrReadFile)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filePath": filePath,
		}).Fatal("Failed to read bytes from file!")
		return os.File{}, nil, err
	}

	log.Info("Finished readOrCreateFile()")
	return *createdOrReadFile, byteValue, nil
}

// GetOrCreateDirectory receives the path to the
// maps directory and creates it if it doesn't exist.
func GetOrCreateDirectory(pathToMapsDirectory string) error {

	log.Info("Entered CreateMapsDirectory()")

	// Create the maps directory:
	err := os.Mkdir(pathToMapsDirectory, 0777)
	if os.IsExist(err) {
		log.Info("The maps directory already exists!")
		return nil
	}
	if err != nil {
		log.WithField("error", err).
			Fatal("failed to create the maps directory!")
		return fmt.Errorf("failed to create the maps directory: %v", err)
	}

	log.Info("Finished GetOrCreateMapsDirectory()")
	return nil
}

// CreateProcessingInfoFile receives a fileNumber and creates a processing info
// file holding the information on which files were processed successfully and which failed.
func CreateProcessingInfoFile(
	logsFilepath string,
	fileNumber int) (*os.File, data.ProcessingInfo, error) {

	log.Info("Entered CreateProcessingInfoFile()")

	// Formatting the processing info file name:
	processingLogName := fmt.Sprintf(logsFilepath+"processed_failed_%v.log", fileNumber)
	processingInfoFile, _, err := ReadOrCreateFile(
		processingLogName,
		os.O_CREATE|os.O_TRUNC|os.O_RDWR,
	)
	if err != nil {
		log.Error("Failed to create the processing info file!")
		return nil, data.ProcessingInfo{}, err
	}

	// This will hold: {"processedFiles": [path, path, path], "failedFiles": [path, path, path]}
	processingInfoStruct := data.NewProcessingInfo()
	// SaveProcessingInfo(&processingInfoFile, processingInfoStruct)

	log.Infof("Created and saved the %v", processingLogName)
	log.Info("Finished CreateProcessingInfoFile()")

	return &processingInfoFile, processingInfoStruct, nil
}

// SaveProcessingInfo receives a file and marshals/writes processingInfoStruct into the file.
func SaveProcessingInfo(
	processingInfoFile *os.File,
	processingInfoStruct data.ProcessingInfo) {

	log.Info("Entered SaveProcessingInfo()")

	processingInfoBytes, err := json.Marshal(processingInfoStruct)
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to marshal processingInfoStruct that is used to create processing.log")
	}

	// Writing processingInfo to the file:
	_, err = processingInfoFile.Write(processingInfoBytes)
	if err != nil {
		log.WithField("error", err).Fatal("Failed to save the processingInfoFile")
	}

	log.Info("Finished SaveProcessingInfo()")
}

// CreatePackageSummaryFile receives packageSummaryStruct and fileNumber
// then saves the package summary file onto the drive.
func CreatePackageSummaryFile(
	absolutePathOutputDirectory string,
	packageSummaryStruct data.PackageSummary,
	fileNumber int) error {
	log.Info("Entered CreatePackageSummaryFile()")

	packageSummaryFilename := fmt.Sprintf("package_summary_%v.json", fileNumber)
	packageAbsolutePath := filepath.Join(absolutePathOutputDirectory, packageSummaryFilename)
	packageSummaryFile, _, err := ReadOrCreateFile(packageAbsolutePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR)
	if err != nil {
		log.Error("Failed to create the package summary file!")
		return err
	}

	packageSummaryBytes, err := json.Marshal(packageSummaryStruct)
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to marshal packageSummaryStruct")
		return fmt.Errorf("Failed to marshal packageSummaryStruct: %v", err)
	}
	_, err = packageSummaryFile.Write(packageSummaryBytes)
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to save the packageSummaryFile")
		return fmt.Errorf("Failed to save the packageSummaryFile: %v", err)
	}

	err = packageSummaryFile.Close()
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to cloes the packageSummaryFile")
		return fmt.Errorf("Failed to close the packageSummaryFile: %v", err)
	}

	log.Info("Finished CreatePackageSummaryFile()")
	return nil
}

// UnmarshalLocaleMapping wraps around unmarshalLocaleFile and returns
// an empty map[string]interface{} if it fails to unmarshal the original locale mapping file.
func UnmarshalLocaleMapping(pathToMappingFile string) map[string]interface{} {

	log.Info("Entered unmarshalLocaleMapping()")

	localizedMapping := make(map[string]interface{})

	if !UnmarshalJsonFile(pathToMappingFile, &localizedMapping) {
		log.WithField("pathToMappingFile", pathToMappingFile).
			Error("Failed to open and unmarshal the mapping file!")
		return localizedMapping
	}

	log.Info("Finished unmarshalLocaleMapping()")

	return localizedMapping
}

// UnmarshalJsonFile deals with every possible opening and unmarshalling
// problem that might occur when unmarshalling a localization file
// supplied by: https://github.com/Kaszanas/SC2MapLocaleExtractor
func UnmarshalJsonFile(
	filepath string,
	mapToPopulate *map[string]interface{}) bool {
	log.Info("Entered unmarshalJsonFile()")

	var file, err = os.Open(filepath)
	if err != nil {
		log.WithField("fileError", err.Error()).
			Info("Failed to open the JSON file.")
		return false
	}
	defer file.Close()

	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		log.WithField("readError", err.Error()).
			Info("Failed to read the JSON file.")
		return false
	}

	err = json.Unmarshal([]byte(jsonBytes), &mapToPopulate)
	if err != nil {
		log.WithField("jsonMarshalError", err.Error()).
			Info("Could not unmarshal the JSON file.")
	}

	log.Info("Finished unmarshalJsonFile()")
	return true
}

// SaveFileToDrive is a helper function that takes
// the json string of a StarCraft II replay and writes it to drive.
func SaveFileToDrive(
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
