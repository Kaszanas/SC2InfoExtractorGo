package settings

import (
	"path/filepath"
)

// TODO: This should be set in some sort of settings:
// This should also use environment variables:
var TEST_LOGS_DIR = "./test_files/test_logs/"
var TEST_LOCALIZATION_FILE_PATH = "./test_files/test_map_mapping/output.json"

var TEST_INPUT_DIR = "./test_files/test_replays"
var TEST_OUTPUT_DIR = "./test_files/test_replays_output/"
var TEST_PROCESSED_FAILED_LOG = TEST_LOGS_DIR + "processed_failed_0.log"

// REVIEW: Is it better to have an environment variable for the workspace directory?
// Or is it better to have that in a .env file?
// How to load a .env file? that is outside of the package?
func GetWorkspaceDirectory() (string, error) {

	workspace, err := filepath.Abs("../")
	if err != nil {
		return "", err
	}

	return workspace, nil
}

func GetTestLogsDirectory() (string, error) {
	workspace, err := GetWorkspaceDirectory()
	if err != nil {
		return "", err
	}

	logsDir := filepath.Join(workspace, "test_files/test_logs")

	return logsDir, nil
}

func GetTestLocalizationFilePath() (string, error) {
	workspace, err := GetWorkspaceDirectory()
	if err != nil {
		return "", err
	}

	localizationFilePath := filepath.Join(workspace, "test_files/test_map_mapping/output.json")

	return localizationFilePath, nil
}

func GetTestInputDirectory() (string, error) {
	workspace, err := GetWorkspaceDirectory()
	if err != nil {
		return "", err
	}

	inputDir := filepath.Join(workspace, "test_files/test_replays")

	return inputDir, nil
}

func GetTestOutputDirectory() (string, error) {
	workspace, err := GetWorkspaceDirectory()
	if err != nil {
		return "", err
	}

	outputDir := filepath.Join(workspace, "test_files/test_replays_output")

	return outputDir, nil
}

func GetTestProcessedFailedLog() (string, error) {
	workspace, err := GetWorkspaceDirectory()
	if err != nil {
		return "", err
	}

	processedFailedLog := filepath.Join(workspace, "test_files/test_logs/processed_failed_0.log")

	return processedFailedLog, nil
}
