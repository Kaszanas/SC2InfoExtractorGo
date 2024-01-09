package dataproc

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils_test"
	log "github.com/sirupsen/logrus"
)

var TEST_INPUT_REPLAYPACK_DIR = ""
var TEST_BYPASS_THESE = []string{}

func TestPipelineWrapper(t *testing.T) {

	flags, chunksOfFiles, logFile, packageToZip, compressionMethod, testLocalizationFilePath, testProcessedFailedlog, testLogsDir, testOutputDir, lenSliceOfFiles := utils_test.SetTestCLIFlags(t)

	localizedMapsMap := map[string]interface{}(nil)
	localizedMapsMap = utils.UnmarshalLocaleMapping(testLocalizationFilePath)
	if localizedMapsMap == nil {
		cleanupOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, false, false)
		if !cleanupOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		log.Fatalf("Test Failed! Could not unmarshall the localization mapping file.")
	}

	PipelineWrapper(
		chunksOfFiles,
		packageToZip,
		localizedMapsMap,
		compressionMethod,
		flags,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]interface{}(nil)
	utils.UnmarshalJsonFile(testProcessedFailedlog, &logFileMap)

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
	if sumProcessed != lenSliceOfFiles {
		cleanukOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, false, false)
		if !cleanukOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		t.Fatalf("Test Failed! input files and processed_failed information mismatch.")
	}

	// Read and verify if the created summaries contain the same count as the processed files
	var summary datastruct.PackageSummary
	unmarshalOk := unmarshalSummaryFile(testOutputDir+"/package_summary_0.json", &summary)
	if !unmarshalOk {
		cleanukOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, false, false)
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
		cleanukOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, false, false)
		if !cleanukOk {
			t.Fatalf("Test Failed! %s", reason)
		}
		t.Fatalf("Test Failed! gameVersion histogram count is different from number of processed files.")
	}

	cleanukOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, false, false)
	if !cleanukOk {
		t.Fatalf("Test Failed! %s", reason)
	}

}

func TestPipelineWrapperMultiple(t *testing.T) {

	if TEST_INPUT_REPLAYPACK_DIR == "" {
		t.SkipNow()
	}

	files, err := os.ReadDir(TEST_INPUT_REPLAYPACK_DIR)
	if err != nil {
		t.Fatalf("Could not read the TEST_INPUT_REPLAYPACK_DIR")
		log.Fatal(err)
	}

	for _, file := range files {
		file := file
		if file.IsDir() {
			filename := file.Name()
			if !contains(TEST_BYPASS_THESE, filename) {
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

}

func testPipelineWrapperWithDir(replayInputPath string, replaypackName string) (bool, string) {

	flags, chunksOfFiles, logFile, packageToZip, compressionMethod, testLocalizationFilePath, testProcessedFailedlog, testLogsDir, testOutputDir, lenSliceOfFiles := utils_test.SetTestCLIFlags(t)

	localizedMapsMap := map[string]interface{}(nil)
	localizedMapsMap = utils.UnmarshalLocaleMapping(testLocalizationFilePath)
	if localizedMapsMap == nil {
		return false, "Could not unmarshall the localization mapping file."
	}

	PipelineWrapper(
		chunksOfFiles,
		packageToZip,
		localizedMapsMap,
		compressionMethod,
		flags,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]interface{}(nil)
	unmarshalOk := utils.UnmarshalJsonFile(testProcessedFailedlog, &logFileMap)
	if !unmarshalOk {
		cleanupOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, true, true)
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
	if sumProcessed != lenSliceOfFiles {
		cleanupOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, true, true)
		if !cleanupOk {
			return false, reason
		}
		return false, "input files and processed_failed information mismatch."
	}

	// Read and verify if the created summaries contain the same count as the processed files
	var summary datastruct.PackageSummary
	unmarshalOk = unmarshalSummaryFile(testOutputDir+"\\package_summary_0.json", &summary)
	if !unmarshalOk {
		cleanupOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, true, true)
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
		cleanupOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, true, true)
		if !cleanupOk {
			return false, reason
		}
		return false, "gameVersion histogram count is different from number of processed files."
	}

	cleanupOk, reason := cleanup(testProcessedFailedlog, testLogsDir, testOutputDir, logFile, true, true)
	if !cleanupOk {
		return false, reason
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

func unmarshalSummaryFile(pathToSummaryFile string, mappingToPopulate *datastruct.PackageSummary) bool {

	log.Info("Entered unmarshalJsonFile()")

	var file, err = os.Open(pathToSummaryFile)
	if err != nil {
		log.WithField("fileError", err.Error()).Info("Failed to open the JSON file.")
		return false
	}
	defer file.Close()

	jsonBytes, err := io.ReadAll(file)
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
