package dataproc

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	settings "github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/chunk_utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

var TEST_BYPASS_THESE_DIRS = []string{}

// KNOWN_UNSUPPORTED_REGION_REPLAYS lists replays recorded on an unsupported SC2
// region (e.g. "Public Test"/PTR), for which the game client's own depot cannot
// provide a map download URL. The pipeline gracefully falls back to an already-cached
// copy of the map when one is available (see
// dataproc/sc2_map_processing.GetDependencyURLsAndHashFromReplayData), so this
// exclusion is only a safety net for runs where the map isn't cached yet.
// TODO: remove once confirmed reliably cached, or replace with a supported-region replay.
var KNOWN_UNSUPPORTED_REGION_REPLAYS = map[string][]string{
	"2017_WCS_Montreal": {"a03dd9e4562845d4a8e986bfd929fd35.SC2Replay"},
}

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

	// Log files are only closed once every subtest has finished (see the
	// t.Cleanup below), instead of at the end of each individual subtest.
	// Closing per-subtest raced with logrus' global, process-wide logger:
	// something could still be logging in the brief window between one
	// subtest's Close() and the next subtest's SetLogging() reassigning the
	// output, producing spurious "Failed to write to log ... file already
	// closed" errors on stderr.
	var openLogFiles []*os.File
	t.Cleanup(func() {
		for _, f := range openLogFiles {
			if err := f.Close(); err != nil {
				t.Errorf("Test Failed! Could not close log file %s: %v", f.Name(), err)
			}
		}
	})

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
						err = os.MkdirAll(thisTestOutputDir, 0755)
						if err != nil {
							t.Fatal("Test Failed! Could not create output directory for test!")
						}
					}

					logFlags := utils.LogFlags{
						LogLevelValue: datastruct.Info,
						LogPath:       thisTestOutputDir,
					}

					logFile, logOk := utils.SetLogging(
						thisTestOutputDir,
						int(logFlags.LogLevelValue),
					)
					if !logOk {
						t.Fatal("Test Failed! Could not perform SetLogging.")
					}
					openLogFiles = append(openLogFiles, logFile)

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

// excludeKnownFilenames removes files whose base name is present in
// excludedFilenames from the given slice of file paths.
func excludeKnownFilenames(filePaths []string, excludedFilenames []string) []string {
	filtered := make([]string, 0, len(filePaths))
	for _, filePath := range filePaths {
		if contains(excludedFilenames, filepath.Base(filePath)) {
			continue
		}
		filtered = append(filtered, filePath)
	}
	return filtered
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
		Debug("Entered testPipelineWrapperWithDir()")

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

	if knownUnsupported, ok := KNOWN_UNSUPPORTED_REGION_REPLAYS[replaypackName]; ok {
		sliceOfFiles = excludeKnownFilenames(sliceOfFiles, knownUnsupported)
	}

	chunksOfFiles, getOk := chunk_utils.GetChunks(sliceOfFiles, 0)
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
		InputDirectory:         replayInputPath,
		OutputDirectory:        thisTestOutputDir,
		OnlyDependencyDownload: false,
		// The shared dependencies/ directory is pre-seeded (via `make
		// fetch_test_fixtures`) from the published kaszanas/sc2reset_maps_mods
		// image, which already contains every map/dependency the historical
		// replay corpus needs, including ones whose original depot (e.g. the
		// since-shut-down CN depot) no longer exists. Tests must not fall back
		// to live downloads from Blizzard's servers: doing so is unnecessary
		// load against a third party, and any map genuinely missing from the
		// pre-seeded cache cannot be fetched live anyway.
		SkipDependencyDownload:     true,
		DependencyDirectory:        "../dependencies/",
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

	// Auxiliary files will be placed in the same directory as the log file:
	foreignToEnglishMappingFilepath := logFlags.LogPath + "map_foreign_to_english_mapping.json"

	foreignToEnglishMapping := downloader.DependencyDownloaderPipeline(
		sliceOfFiles,
		foreignToEnglishMappingFilepath,
		flags,
	)

	PipelineWrapper(
		chunksOfFiles,
		packageToZip,
		compressionMethod,
		foreignToEnglishMapping,
		flags,
	)

	// Read and verify if the processed_failed information contains the same count of files processed as the output
	logFileMap := map[string]any(nil)
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
			true,
			true)
		if err != nil {
			return false, reason
		}
	}

	return true, ""
}

// pipelineTestCleanup is a helper function to clean up the test output directory.
// The log file itself is closed by the caller (TestPipelineWrapperMultiple's
// t.Cleanup, once every subtest has finished) — closing it here too would
// double-close it.
func pipelineTestCleanup(
	processedFailedPath string,
	testOutputPath string,
	deleteOutputDir bool,
	deleteLogsFilepath bool,
) (string, error) {

	// err := os.Remove(processedFailedPath)
	// if err != nil {
	// 	return false, "Cannot delete processed_failed file."
	// }

	err := os.Remove(testOutputPath + "main_log.log")
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

	log.Debug("Entered unmarshalSummaryFile()")

	var file, err = os.Open(pathToSummaryFile)
	if err != nil {
		return "Failed to open the JSON file.", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.WithField("error", err).Error("Failed to close JSON file.")
		}
	}()

	jsonBytes, err := io.ReadAll(file)
	if err != nil {
		return "Failed to read the JSON file.", err
	}

	err = json.Unmarshal([]byte(jsonBytes), &mappingToPopulate)
	if err != nil {
		return "Could not unmarshal the JSON file.", err
	}

	log.Debug("Finished unmarshalSummaryFile()")
	return "", nil
}
