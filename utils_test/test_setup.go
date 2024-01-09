package utils_test

import (
	"os"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func SetTestCLIFlags(t *testing.T) (utils.CLIFlags, [][]string, *os.File, bool, uint16, string, string, string, string, int) {
	testInputDir, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Could not get the test input directory.")
	}
	log.WithField("testInputDir", testInputDir).Info("Input dir was set.")

	testLogsDir, err := settings.GetTestLogsDirectory()
	if err != nil {
		t.Fatalf("Could not get the test logs directory.")
	}
	log.WithField("testLogsDir", testLogsDir).Info("Logs dir was set.")

	testLocalizationFilePath, err := settings.GetTestLocalizationFilePath()
	if err != nil {
		t.Fatalf("Could not get the test localization file path.")
	}
	log.WithField("testLocalizationFilePath", testLocalizationFilePath).
		Info("Localization file path was set.")

	testProcessedFailedlog, err := settings.GetTestProcessedFailedLog()
	if err != nil {
		t.Fatalf("Could not get the test processed_failed log.")
	}
	log.WithField("testProcessedFailedlog", testProcessedFailedlog).
		Info("Processed failed log path was set.")

	testOutputDir, err := settings.GetTestOutputDirectory()
	if err != nil {
		t.Fatalf("Could not get the test output directory.")
	}
	log.WithField("testOutputDir", testOutputDir).Info("Output dir was set.")

	sliceOfFiles := utils.ListFiles(testInputDir, ".SC2Replay")
	if len(sliceOfFiles) < 1 {
		t.Fatalf("Could not detect test files! Verify if they exist.")
	}
	log.WithField("n_files", len(sliceOfFiles)).Info("Number of detected files.")

	chunks, getOk := utils.GetChunksOfFiles(sliceOfFiles, 0)
	if !getOk {
		t.Fatalf("Test Failed! Could not produce chunks of files!")
	}

	logFile, logOk := utils.SetLogging(testLogsDir, 3)
	if !logOk {
		t.Fatalf("Test Failed! Could not perform SetLogging.")
	}

	// Create dummy CLI flags:
	gameModeCheckFlag := 0
	flags := utils.CLIFlags{
		InputDirectory:             testInputDir,
		OutputDirectory:            testOutputDir,
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
			LogPath:  testLogsDir,
		},
		CPUProfilingPath: "",
	}

	packageToZip := true
	compressionMethod := uint16(8)

	return flags, chunks, logFile, packageToZip, compressionMethod, testLocalizationFilePath, testProcessedFailedlog, testLogsDir, testOutputDir, len(sliceOfFiles)
}
