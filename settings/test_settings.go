package settings

import (
	"path/filepath"
)

// REVIEW: Is it better to have an environment variable for the workspace directory?
// Or is it better to have that in a .env file?
// How to load a .env file? that is outside of the package?
func GetWorkspaceDirectory() (string, error) {

	// REVIEW: Will this consistently point to the workspace?
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
	logsDirectory, err := GetTestLogsDirectory()
	if err != nil {
		return "", err
	}

	processedFailedLog := filepath.Join(logsDirectory, "processed_failed_0.log")

	return processedFailedLog, nil
}
