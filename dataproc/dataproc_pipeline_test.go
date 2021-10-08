package dataproc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func TestPipelineWrapper(t *testing.T) {

	testOutputPath := "../test_files/test_replays_output/"
	// testOutputPath := t.TempDir()

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

	packageToZip := true
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

	// Read and verify if the created summaries contain the same count as the processed files
	var summary datastruct.PackageSummary
	unmarshalOk := unmarshalSummaryFile(testOutputPath+"\\package_summary_0.json", &summary)
	if !unmarshalOk {
		t.Fatalf("Test Failed! Cannot read summary file.")
	}

	gameVersionCount := 0
	for _, value := range summary.Summary.GameVersions {
		gameVersionCount += int(value)
	}

	if gameVersionCount != len(sliceOfFiles) {
		t.Fatalf("Test Failed! gameVersion histogram count is different from number of processed files.")
	}

	t.Cleanup(func() {
		cleanup(processedFailedPath, logsFilepath, testOutputPath, logFile, t)
	})

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

func cleanup(processedFailedPath string, logsFilepath string, testOutputPath string, logFile *os.File, t *testing.T) {
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

	filesToClean := utils.ListFiles(testOutputPath, "")
	for _, file := range filesToClean {
		err = os.Remove(file)
		if err != nil {
			t.Fatalf("Test Failed! Cannot delete output files.")
		}
	}
}

func unmarshalSummaryFile(pathToSummaryFile string, mappingToPopulate *datastruct.PackageSummary) bool {

	log.Info("Entered unmarshalJsonFile()")

	var file, err = os.Open(pathToSummaryFile)
	if err != nil {
		log.WithField("fileError", err.Error()).Info("Failed to open the JSON file.")
		return false
	}
	defer file.Close()

	jsonBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithField("readError", err.Error()).Info("Failed to read the JSON file.")
		return false
	}

	err = json.Unmarshal([]byte(jsonBytes), &mappingToPopulate)
	if err != nil {
		log.WithField("jsonMarshalError", err.Error()).Info("Could not unmarshal the JSON file.")
	}

	log.Info("Finished unmarshalJsonFile()")

	return true
}
