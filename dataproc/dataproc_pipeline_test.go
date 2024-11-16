package dataproc

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	settings "github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/chunk_utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

var TEST_BYPASS_THESE_DIRS = []string{}

// TestPipelineWrapperSingle is a test function to test the pipeline wrapper
// on all of the replaypack directories in the test input directory.
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

	// dirContents = []fs.DirEntry{dirContents[3]}

	for _, maybeDir := range dirContents {
		if maybeDir.IsDir() {
			dirName := maybeDir.Name()
			if !contains(TEST_BYPASS_THESE_DIRS, dirName) {
				absoluteTestReplayDir := filepath.Join(testInputDir, dirName)
				t.Run(dirName, func(t *testing.T) {
					// WARNING: Cannot run tests in parallel
					// because of the downloading logic, it reads from a common
					// maps directory:
					// t.Parallel()

					// This and all below is done here because
					// the logging should be set before the test starts.
					// Otherwise part of the logs will be saved in the previous tests log.
					testOutputDir, err := settings.GetTestOutputDirectory()
					if err != nil {
						t.Fatal("Test Failed! Could not get the test output directory.")
					}

					thisTestOutputDir := testOutputDir + "/" + dirName + "/"
					log.WithField("thisTestOutputDir", thisTestOutputDir).
						Info("Defined a path for the output of the test.")
					if _, err := os.Stat(thisTestOutputDir); os.IsNotExist(err) {
						log.WithField("thisTestOutputDir", thisTestOutputDir).
							Info("Test output dir does not exist, attempting to create.")
						err = os.Mkdir(thisTestOutputDir, 0755)
						if err != nil {
							t.Fatal("Test Failed! Could not create output directory for test!")
						}
					}

					logFlags := utils.LogFlags{
						LogLevelValue: datastruct.Info,
						LogPath:       thisTestOutputDir,
					}

					logFile, logOk := utils.SetLogging(thisTestOutputDir, int(logFlags.LogLevelValue))
					defer logFile.Close()
					if !logOk {
						t.Fatal("Test Failed! Could not perform SetLogging.")
					}

					testOk, reason := testPipelineWrapperWithDir(
						thisTestOutputDir,
						absoluteTestReplayDir,
						dirName,
						logFile,
						logFlags,
						removeTestOutputs)
					if !testOk {
						t.Fatalf("Test Failed! %s", reason)
					}
				})
			}
		}
	}

}

// testPipelineWrapperWithDir is a helper function to test the pipeline wrapper
// on a single replaypack directory.
func testPipelineWrapperWithDir(
	thisTestOutputDir string,
	replayInputPath string,
	replaypackName string,
	logFile *os.File,
	logFlags utils.LogFlags,
	removeTestOutputs bool,
) (bool, string) {

	log.WithFields(log.Fields{"testOutputDir": thisTestOutputDir}).
		Info("Entered testPipelineWrapperWithDir()")

	// TODO: This should be refactored, new hybrid approach should be applied
	// https://github.com/Kaszanas/SC2InfoExtractorGo/issues/49
	testLocalizationFilePath, err := settings.GetTestLocalizationFilePath()
	if err != nil {
		return false, "Could not get the test localization file path."
	}
	log.WithField("testLocalizationFilePath", testLocalizationFilePath).
		Info("Got test localization filepath from settings.")

	sliceOfFiles, err := file_utils.ListFiles(replayInputPath, ".SC2Replay")
	if err != nil {
		return false, "Could not get the list of files."
	}

	chunksOfFiles, getOk := chunk_utils.GetChunksOfFiles(sliceOfFiles, 0)
	if !getOk {
		return false, "Could not produce chunks of files!"
	}
	log.WithFields(log.Fields{
		"n_files":      len(sliceOfFiles),
		"sliceOfFiles": sliceOfFiles}).Info("Got files to test.")

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
		LogFlags:                   logFlags,
		CPUProfilingPath:           "",
	}

	packageToZip := true
	compressionMethod := uint16(8)

	PipelineWrapper(
		chunksOfFiles,
		packageToZip,
		compressionMethod,
		settings.MapsDirectoryPath,
		settings.DownloadedMapsForReplaysFilepath,
		settings.ForeignToEnglishMappingFilepath,
		flags,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]interface{}(nil)
	processedFailedPath := thisTestOutputDir + "processed_failed_0.log"
	err = file_utils.UnmarshalJsonFile(processedFailedPath, &logFileMap)
	if err != nil {
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
	if failedToProcessCount > 0 {
		return false, "Failed to process count more than 0"
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
	var summary persistent_data.PackageSummary
	pathToSummaryFile := thisTestOutputDir + "/" + "package_summary_0.json"
	log.WithField("pathToSummaryFile", pathToSummaryFile).
		Info("Set the path to the summary file.")
	reason, err := unmarshalSummaryFile(
		pathToSummaryFile,
		&summary)
	if err != nil {
		log.WithField("error", err.Error()).
			Info(reason)
		return false, reason
	}

	histogramGameVersionCount := 0
	for _, value := range summary.Summary.GameVersions {
		histogramGameVersionCount += int(value)
	}

	if histogramGameVersionCount != processedFilesCount {
		return false,
			"gameVersion histogram count is different from number of processed files."
	}

	if removeTestOutputs {
		reason, err = pipelineTestCleanup(
			processedFailedPath,
			thisTestOutputDir,
			logFile,
			true,
			true)
		if err != nil {
			return false, reason
		}
	}

	return true, ""
}

// pipelineTestCleanup is a helper function to clean up the test output directory.
func pipelineTestCleanup(
	processedFailedPath string,
	testOutputPath string,
	logFile *os.File,
	deleteOutputDir bool,
	deleteLogsFilepath bool) (string, error) {

	// err := os.Remove(processedFailedPath)
	// if err != nil {
	// 	return false, "Cannot delete processed_failed file."
	// }

	err := logFile.Close()
	if err != nil {
		return "Cannot close the main_log file.", err
	}

	err = os.Remove(testOutputPath + "main_log.log")
	if err != nil {
		return "Cannot delete main_log file.", err
	}

	if deleteOutputDir {
		err = os.RemoveAll(testOutputPath)
		if err != nil {
			return "Cannot delete output path.", err
		}
	} else {
		filesToClean, err := file_utils.ListFiles(testOutputPath, "")
		if err != nil {
			return "Cannot get the files in the cleanup directory.", err
		}

		for _, file := range filesToClean {
			err = os.Remove(file)
			if err != nil {
				return "Cannot delete output files.", err
			}
		}
	}

	return "", nil
}

func unmarshalSummaryFile(
	pathToSummaryFile string,
	mappingToPopulate *persistent_data.PackageSummary) (string, error) {

	log.Info("Entered unmarshalSummaryFile()")

	var file, err = os.Open(pathToSummaryFile)
	if err != nil {
		return "Failed to open the JSON file.", err
	}
	defer file.Close()

	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		return "Failed to read the JSON file.", err
	}

	err = json.Unmarshal([]byte(jsonBytes), &mappingToPopulate)
	if err != nil {
		return "Could not unmarshal the JSON file.", err
	}

	log.Info("Finished unmarshalSummaryFile()")

	return "", nil
}
