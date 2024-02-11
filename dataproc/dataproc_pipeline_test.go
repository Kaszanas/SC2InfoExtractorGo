package dataproc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	settings "github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

var TEST_BYPASS_THESE_DIRS = []string{}

func TestPipelineWrapperMultiple(t *testing.T) {

	removeTestOutputs := settings.DELETE_TEST_OUTPUT

	testInputDir, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Could not get the test input directory.")
	}
	log.WithField("testInputDir", testInputDir).Info("Input dir was set.")

	dirContents, err := os.ReadDir(testInputDir)
	if err != nil {
		t.Fatalf("Could not get the test directory contents.")
		log.Fatal(err)
	}

	dirContents = []fs.DirEntry{dirContents[1]}

	for _, maybeDir := range dirContents {
		if maybeDir.IsDir() {
			dirName := maybeDir.Name()
			if !contains(TEST_BYPASS_THESE_DIRS, dirName) {
				absoluteReplayDir := filepath.Join(testInputDir, dirName)
				t.Run(dirName, func(t *testing.T) {
					// t.Parallel()
					testOk, reason := testPipelineWrapperWithDir(
						absoluteReplayDir,
						dirName,
						removeTestOutputs)
					if !testOk {

						t.Fatalf("Test Failed! %s", reason)
					}
				})
			}
		}
	}

}

func testPipelineWrapperWithDir(
	replayInputPath string,
	replaypackName string,
	removeTestOutputs bool) (bool, string) {

	testOutputDir, err := settings.GetTestOutputDirectory()
	if err != nil {
		return false, "Could not get the test output directory."
	}
	log.WithField("testOutputDir", testOutputDir).
		Info("Got test output directory from settings.")

	testLocalizationFilePath, err := settings.GetTestLocalizationFilePath()
	if err != nil {
		return false, "Could not get the test localization file path."
	}
	log.WithField("testLocalizationFilePath", testLocalizationFilePath).
		Info("Got test localization filepath from settings.")

	sliceOfFiles := utils.ListFiles(replayInputPath, ".SC2Replay")
	chunksOfFiles, getOk := utils.GetChunksOfFiles(sliceOfFiles, 0)
	if !getOk {
		return false, "Could not produce chunks of files!"
	}
	log.WithFields(log.Fields{
		"n_files":      len(sliceOfFiles),
		"sliceOfFiles": sliceOfFiles}).Info("Got files to test.")

	thisTestOutputDir := testOutputDir + "/" + replaypackName + "/"
	log.WithField("thisTestOutputDir", thisTestOutputDir).
		Info("Defined a path for the output of the test.")
	err = os.Mkdir(thisTestOutputDir, 0755)
	if err != nil {
		return false, "Could not create output directory for test!"
	}

	logFile, logOk := utils.SetLogging(thisTestOutputDir, 3)
	if !logOk {
		return false, "Could not perform SetLogging."
	}

	// REVIEW: Hardcoded flags for test? I suppose that these
	// should come from a specific test case.
	// Create dummy CLI flags:
	gameModeCheckFlag := 0
	flags := utils.CLIFlags{
		InputDirectory:             replayInputPath,
		OutputDirectory:            thisTestOutputDir,
		NumberOfThreads:            1,
		NumberOfPackages:           1,
		PerformIntegrityCheck:      true,
		PerformValidityCheck:       false,
		PerformCleanup:             true,
		PerformPlayerAnonymization: false,
		PerformChatAnonymization:   false,
		PerformFiltering:           false,
		FilterGameMode:             gameModeCheckFlag,
		LocalizationMapFile:        testLocalizationFilePath,
		LogFlags: utils.LogFlags{
			LogLevel: 5,
			LogPath:  thisTestOutputDir,
		},
		CPUProfilingPath: "",
	}

	packageToZip := true
	compressionMethod := uint16(8)

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
	processedFailedPath := thisTestOutputDir + "processed_failed_0.log"
	unmarshalOk := utils.UnmarshalJsonFile(processedFailedPath, &logFileMap)
	if !unmarshalOk {
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
		return false, "input files and processed_failed information mismatch."
	}

	// Read and verify if the created summaries contain the same count as the processed files
	var summary datastruct.PackageSummary
	unmarshalOk = unmarshalSummaryFile(
		thisTestOutputDir+"\\package_summary_0.json",
		&summary)
	if !unmarshalOk {
		return false, "Cannot read summary file."
	}

	histogramGameVersionCount := 0
	for _, value := range summary.Summary.GameVersions {
		histogramGameVersionCount += int(value)
	}

	if histogramGameVersionCount != processedFilesCount {
		return false, "gameVersion histogram count is different from number of processed files."
	}

	if removeTestOutputs {
		cleanupOk, reason := pipelineTestCleanup(
			processedFailedPath,
			thisTestOutputDir,
			logFile,
			true,
			true)
		if !cleanupOk {
			return false, reason
		}
	}

	return true, ""
}

func pipelineTestCleanup(
	processedFailedPath string,
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

	err = os.Remove(testOutputPath + "main_log.log")
	if err != nil {
		return false, "Cannot delete main_log file."
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

func unmarshalSummaryFile(
	pathToSummaryFile string,
	mappingToPopulate *datastruct.PackageSummary) bool {

	log.Info("Entered unmarshalJsonFile()")

	var file, err = os.Open(pathToSummaryFile)
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

	err = json.Unmarshal([]byte(jsonBytes), &mappingToPopulate)
	if err != nil {
		log.WithField("jsonMarshalError", err.Error()).
			Info("Could not unmarshal the JSON file.")
	}

	log.Info("Finished unmarshalJsonFile()")

	return true
}
