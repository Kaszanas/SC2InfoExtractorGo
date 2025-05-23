package persistent_data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

// ProcessingInfo is a structure holding information that
// is used to create processing.log, which is anonymizedPlayers
// in a persistent map from toon to unique integer,
// slice of processed files so that there is a state of all of the processed files.
type ProcessingInfo struct {
	ProcessedFiles  []string            `json:"processedFiles"`
	FailedToProcess []map[string]string `json:"failedToProcess"`
}

// NewProcessingInfo returns empty ProcessingIngo struct.
func NewProcessingInfo() ProcessingInfo {
	return ProcessingInfo{
		ProcessedFiles:  make([]string, 0),
		FailedToProcess: make([]map[string]string, 0),
	}
}

// AddToProcessed adds a replay file path to the list of processed files.
func (processingInfo *ProcessingInfo) AddToProcessed(replayFilePath string) {
	replayFileNameAndExtension := filepath.Base(replayFilePath)

	processingInfo.ProcessedFiles = append(
		processingInfo.ProcessedFiles,
		replayFileNameAndExtension,
	)
}

// AddToFailed adds a replay file path to the list of failed files.
// Includes a reason for failure.
func (processingInfo *ProcessingInfo) AddToFailed(
	replayFilePath string,
	reason string,
) {
	replayFileNameAndExtension := filepath.Base(replayFilePath)

	processingInfo.FailedToProcess = append(
		processingInfo.FailedToProcess, map[string]string{
			"fileName": replayFileNameAndExtension,
			"reason":   reason,
		})
}

// CreateProcessingInfoFile receives a fileNumber and creates a processing info
// file holding the information on which files were processed successfully and which failed.
func CreateProcessingInfoFile(
	logsFilepath string,
	fileNumber int,
) (*os.File, ProcessingInfo, error) {

	log.Debug("Entered CreateProcessingInfoFile()")

	// Formatting the processing info file name:
	processingLogName := fmt.Sprintf(logsFilepath+"processed_failed_%v.log", fileNumber)
	processingInfoFile, err := file_utils.CreateTruncateFile(
		processingLogName,
	)
	if err != nil {
		log.Error("Failed to create the processing info file!")
		return nil, ProcessingInfo{}, err
	}

	// This will hold: {"processedFiles": [path, path, path], "failedFiles": [path, path, path]}
	processingInfoStruct := NewProcessingInfo()
	// SaveProcessingInfo(&processingInfoFile, processingInfoStruct)

	log.Infof("Created and saved the %v", processingLogName)

	log.Debug("Finished CreateProcessingInfoFile()")
	return &processingInfoFile, processingInfoStruct, nil
}

// SaveProcessingInfoToFile receives a file
// and marshals/writes processingInfoStruct into the file.
func SaveProcessingInfoToFile(
	processingInfoFile *os.File,
	processingInfoStruct ProcessingInfo,
) {

	log.Debug("Entered SaveProcessingInfo()")

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

	log.Debug("Finished SaveProcessingInfo()")
}
