package dataproc

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func TestPipelineWrapper(t *testing.T) {

	testOutputPath := t.TempDir()

	testLocalizationMapFile := "../test_files/test_map_mapping/output.json"
	logsFilepath := "../test_files/test_logs/"

	testReplayDir := "../test_files/test_replays"
	sliceOfFiles := utils.ListFiles(testReplayDir, ".SC2Replay")
	chunks, getOk := utils.GetChunksOfFiles(sliceOfFiles, 0)
	if !getOk {
		t.Fatalf("Test Failed! Could not produce chunks of files!")
	}

	logFile, logOk := utils.SetLogging(logsFilepath, 3)
	if !logOk {
		t.Fatalf("Test Failed! Could not perform SetLogging.")
	}

	packageToZip := false
	integrityCheck := true
	validityCheck := false
	performFilteringBool := false
	gameModeCheckFlag := 0
	performPlayerAnonymization := false
	performChatAnonymization := false
	performCleanup := true
	compressionMethod := uint16(8)
	numberOfThreads := 1

	localizedMapsMap := map[string]interface{}(nil)
	localizedMapsMap = utils.UnmarshalLocaleMapping(testLocalizationMapFile)
	if localizedMapsMap == nil {
		log.Fatalf("Test Failed! Could not unmarshall the localization mapping file.")
	}

	PipelineWrapper(
		testOutputPath,
		chunks,
		packageToZip,
		integrityCheck,
		validityCheck,
		performFilteringBool,
		gameModeCheckFlag,
		performPlayerAnonymization,
		performChatAnonymization,
		performCleanup,
		localizedMapsMap,
		compressionMethod,
		numberOfThreads,
		logsFilepath,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]interface{}(nil)
	processedFailedPath := logsFilepath + "processed_failed_0.log"
	utils.UnmarshalJsonFile(processedFailedPath, &logFileMap)

	failedToProcessEmpty := isNil(logFileMap["failedToProcess"].([]interface{}))
	var failedToProcessCount int
	failedToProcessCount = 0
	if !failedToProcessEmpty {
		failedSlice := []string{}
		for _, v := range logFileMap["failedToProcess"].([]interface{}) {
			failedSlice = append(failedSlice, fmt.Sprint(v))
		}

		failedToProcessCount = len(failedSlice)
	}

	processedFilesEmpty := isNil(logFileMap["processedFiles"].([]interface{}))
	var processedFilesCount int
	processedFilesCount = 0
	if !processedFilesEmpty {
		processedSlice := []string{}
		for _, v := range logFileMap["processedFiles"].([]interface{}) {
			processedSlice = append(processedSlice, fmt.Sprint(v))
		}

		processedFilesCount = len(processedSlice)
	}

	sumProcessed := processedFilesCount + failedToProcessCount
	if sumProcessed != len(sliceOfFiles) {
		t.Fatalf("Test Failed! input files and processed_failed information mismatch.")
	}

	err := os.Remove(processedFailedPath)
	if err != nil {
		t.Fatalf("Test Failed! Cannot delete processed_failed file.")
	}

	err = logFile.Close()
	if err != nil {
		t.Fatalf("Test Failed! Cannot close the main_log file.")
	}

	err = os.Remove(logsFilepath + "main_log.log")
	if err != nil {
		t.Fatalf("Test Failed! Cannot delete main_log file.")
	}

	// Read and verify if the created summaries contain the same count as the processed files

}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
