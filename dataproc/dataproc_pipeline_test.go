package dataproc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

var TEST_LOGS_DIR = "../test_files/test_logs/"
var TEST_LOCALIZATION_FILE_PATH = "../test_files/test_map_mapping/output.json"
var TEST_INPUT_REPLAYPACK_DIR = "F:\\Projects\\EsportDataset\\processing_with_python\\input"
var TEST_INPUT_DIR = "../test_files/test_replays"
var TEST_OUTPUT_DIR = "../test_files/test_replays_output/"
var TEST_PROCESSED_FAILED_LOG = TEST_LOGS_DIR + "processed_failed_0.log"

func TestPipelineWrapper(t *testing.T) {

	sliceOfFiles := utils.ListFiles(TEST_INPUT_DIR, ".SC2Replay")
	chunks, getOk := utils.GetChunksOfFiles(sliceOfFiles, 0)
	if !getOk {
		t.Fatalf("Test Failed! Could not produce chunks of files!")
	}

	logFile, logOk := utils.SetLogging(TEST_LOGS_DIR, 3)
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
	localizedMapsMap = utils.UnmarshalLocaleMapping(TEST_LOCALIZATION_FILE_PATH)
	if localizedMapsMap == nil {
		cleanukOk, reason := cleanup(TEST_PROCESSED_FAILED_LOG, TEST_LOGS_DIR, TEST_OUTPUT_DIR, logFile, false, false)
		if !cleanukOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		log.Fatalf("Test Failed! Could not unmarshall the localization mapping file.")
	}

	PipelineWrapper(
		TEST_OUTPUT_DIR,
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
		TEST_LOGS_DIR,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]interface{}(nil)
	utils.UnmarshalJsonFile(TEST_PROCESSED_FAILED_LOG, &logFileMap)

	var failedToProcessCount int
	failedToProcessCount = 0
	if logFileMap["failedToProcess"] != nil {
		failedSlice := []string{}
		for _, v := range logFileMap["failedToProcess"].([]interface{}) {
			failedSlice = append(failedSlice, fmt.Sprint(v))
		}

		failedToProcessCount = len(failedSlice)
	}

	var processedFilesCount int
	processedFilesCount = 0
	if logFileMap["processedFiles"] != nil {
		processedSlice := []string{}
		for _, v := range logFileMap["processedFiles"].([]interface{}) {
			processedSlice = append(processedSlice, fmt.Sprint(v))
		}

		processedFilesCount = len(processedSlice)
	}

	sumProcessed := processedFilesCount + failedToProcessCount
	if sumProcessed != len(sliceOfFiles) {
		cleanukOk, reason := cleanup(TEST_PROCESSED_FAILED_LOG, TEST_LOGS_DIR, TEST_OUTPUT_DIR, logFile, false, false)
		if !cleanukOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		t.Fatalf("Test Failed! input files and processed_failed information mismatch.")
	}

	// Read and verify if the created summaries contain the same count as the processed files
	var summary datastruct.PackageSummary
	unmarshalOk := unmarshalSummaryFile(TEST_OUTPUT_DIR+"\\package_summary_0.json", &summary)
	if !unmarshalOk {
		cleanukOk, reason := cleanup(TEST_PROCESSED_FAILED_LOG, TEST_LOGS_DIR, TEST_OUTPUT_DIR, logFile, false, false)
		if !cleanukOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		t.Fatalf("Test Failed! Cannot read summary file.")
	}

	gameVersionCount := 0
	for _, value := range summary.Summary.GameVersions {
		gameVersionCount += int(value)
	}

	if gameVersionCount != processedFilesCount {
		cleanukOk, reason := cleanup(TEST_PROCESSED_FAILED_LOG, TEST_LOGS_DIR, TEST_OUTPUT_DIR, logFile, false, false)
		if !cleanukOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		t.Fatalf("Test Failed! gameVersion histogram count is different from number of processed files.")
	}

}

func TestPipelineWrapperMultiple(t *testing.T) {

	files, err := ioutil.ReadDir(TEST_INPUT_REPLAYPACK_DIR)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			filename := file.Name()
			absoluteReplayDir := filepath.Join(TEST_INPUT_REPLAYPACK_DIR, filename)

			t.Run(filename, func(t *testing.T) {
				// t.Parallel()
				testOk, reason := testPipelineWrapperWithDir(absoluteReplayDir, filename)
				if !testOk {
					t.Fatalf("Test Failed! %s", reason)
				}
			})
		}

	}

}

func testPipelineWrapperWithDir(replayInputPath string, replaypackName string) (bool, string) {

	sliceOfFiles := utils.ListFiles(replayInputPath, ".SC2Replay")
	chunks, getOk := utils.GetChunksOfFiles(sliceOfFiles, 0)
	if !getOk {
		return false, "Could not produce chunks of files!"
	}

	thisTestLogsDir := TEST_LOGS_DIR + replaypackName + "/"
	err := os.Mkdir(thisTestLogsDir, 0755)
	if err != nil {
		return false, "Could not create logs directory for test!"
	}

	thisTestOutputDir := TEST_OUTPUT_DIR + replaypackName + "/"
	err = os.Mkdir(thisTestOutputDir, 0755)
	if err != nil {
		return false, "Could not create output directory for test!"
	}

	logFile, logOk := utils.SetLogging(thisTestLogsDir, 3)
	if !logOk {
		return false, "Could not perform SetLogging."
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
	localizedMapsMap = utils.UnmarshalLocaleMapping(TEST_LOCALIZATION_FILE_PATH)
	if localizedMapsMap == nil {
		return false, "Could not unmarshall the localization mapping file."
	}

	PipelineWrapper(
		thisTestOutputDir,
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
		thisTestLogsDir,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]interface{}(nil)
	processedFailedPath := thisTestLogsDir + "processed_failed_0.log"
	unmarshalOk := utils.UnmarshalJsonFile(processedFailedPath, &logFileMap)
	if !unmarshalOk {
		cleanupOk, reason := cleanup(processedFailedPath, thisTestLogsDir, thisTestOutputDir, logFile, true, true)
		if !cleanupOk {
			return false, reason
		}
		return false, "Could not unmrshal processed_failed file."
	}

	var failedToProcessCount int
	failedToProcessCount = 0
	if logFileMap["failedToProcess"] != nil {
		failedSlice := []string{}
		for _, v := range logFileMap["failedToProcess"].([]interface{}) {
			failedSlice = append(failedSlice, fmt.Sprint(v))
		}

		failedToProcessCount = len(failedSlice)
	}

	var processedFilesCount int
	processedFilesCount = 0
	if logFileMap["processedFiles"] != nil {
		processedSlice := []string{}
		for _, v := range logFileMap["processedFiles"].([]interface{}) {
			processedSlice = append(processedSlice, fmt.Sprint(v))
		}

		processedFilesCount = len(processedSlice)
	}

	sumProcessed := processedFilesCount + failedToProcessCount
	if sumProcessed != len(sliceOfFiles) {
		cleanupOk, reason := cleanup(processedFailedPath, thisTestLogsDir, thisTestOutputDir, logFile, true, true)
		if !cleanupOk {
			return false, reason
		}
		return false, "input files and processed_failed information mismatch."
	}

	// Read and verify if the created summaries contain the same count as the processed files
	var summary datastruct.PackageSummary
	unmarshalOk = unmarshalSummaryFile(thisTestOutputDir+"\\package_summary_0.json", &summary)
	if !unmarshalOk {
		cleanupOk, reason := cleanup(processedFailedPath, thisTestLogsDir, thisTestOutputDir, logFile, true, true)
		if !cleanupOk {
			return false, reason
		}
		return false, "Cannot read summary file."
	}

	gameVersionCount := 0
	for _, value := range summary.Summary.GameVersions {
		gameVersionCount += int(value)
	}

	if gameVersionCount != processedFilesCount {
		cleanupOk, reason := cleanup(processedFailedPath, thisTestLogsDir, thisTestOutputDir, logFile, true, true)
		if !cleanupOk {
			return false, reason
		}
		return false, "gameVersion histogram count is different from number of processed files."
	}

	return true, ""

}

func cleanup(
	processedFailedPath string,
	logsFilepath string,
	testOutputPath string,
	logFile *os.File,
	deleteOutputDir bool,
	deleteLogsFilepath bool) (bool, string) {

	// err := os.Remove(processedFailedPath)
	// if err != nil {
	// 	return false, "Cannot delete processed_failed file."
	// }

	err := logFile.Close()
	if err != nil {
		return false, "Cannot close the main_log file."
	}

	if deleteLogsFilepath {
		err = os.RemoveAll(logsFilepath)
		if err != nil {
			return false, "Cannot delete logsFilepath."
		}
	} else {
		err = os.Remove(logsFilepath + "main_log.log")
		if err != nil {
			return false, "Cannot delete main_log file."
		}
	}

	if deleteOutputDir {
		err = os.RemoveAll(testOutputPath)
		if err != nil {
			return false, "Cannot delete output path."
		}
	} else {
		filesToClean := utils.ListFiles(testOutputPath, "")
		for _, file := range filesToClean {
			err = os.Remove(file)
			if err != nil {
				return false, "Cannot delete output files."
			}
		}
	}

	return true, ""
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
