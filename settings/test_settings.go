package settings

import (
	"path/filepath"
)

var DELETE_TEST_OUTPUT = false

// REVIEW: Is it better to have an environment variable for the workspace directory?
// Or is it better to have that in a .env file?
// How to load a .env file? that is outside of the package?
// GetWorkspaceDirectory returns the path to the workspace directory.
func GetWorkspaceDirectory() (string, error) {

	// REVIEW: Will this consistently point to the workspace?
	workspace, err := filepath.Abs("../")
	if err != nil {
		return "", err
	}

	return workspace, nil
}

// GetTestFilesDirectory returns the path to the test_files directory.
func GetTestFilesDirectory() (string, error) {
	workspace, err := GetWorkspaceDirectory()
	if err != nil {
		return "", err
	}

	testFilesDir := filepath.Join(workspace, "test_files")

	return testFilesDir, nil
}

// GetTestLogsDirectory returns the path to the test_logs directory.
func GetTestLogsDirectory() (string, error) {
	testFilesDir, err := GetTestFilesDirectory()
	if err != nil {
		return "", err
	}

	logsDir := filepath.Join(testFilesDir, "test_logs")

	return logsDir, nil
}

// GetTestLocalizationFilePath returns the path to the test_map_mapping/output.json file.
func GetTestLocalizationFilePath() (string, error) {
	testFilesDir, err := GetTestFilesDirectory()
	if err != nil {
		return "", err
	}

	localizationFilePath := filepath.Join(testFilesDir, "test_map_mapping/output.json")

	return localizationFilePath, nil
}

// GetTestInputDirectory returns the path to the test_replays directory.
func GetTestInputDirectory() (string, error) {
	testFilesDir, err := GetTestFilesDirectory()
	if err != nil {
		return "", err
	}

	inputDir := filepath.Join(testFilesDir, "test_replays")

	return inputDir, nil
}

// GetTestOutputDirectory returns the path to the test_replays_output directory.
func GetTestOutputDirectory() (string, error) {
	testFilesDir, err := GetTestFilesDirectory()
	if err != nil {
		return "", err
	}

	outputDir := filepath.Join(testFilesDir, "test_replays_output")

	return outputDir, nil
}

// GetTestProcessedFailedLog returns the path to the processed_failed log file.
func GetTestProcessedFailedLog() (string, error) {
	logsDirectory, err := GetTestLogsDirectory()
	if err != nil {
		return "", err
	}

	// TODO: This might change, if there will be more logging files required.
	processedFailedLog := filepath.Join(logsDirectory, "processed_failed_0.log")

	return processedFailedLog, nil
}

// GetProfilerReportPath returns the path to the profiler report file.
func GetProfilerReportPath() (string, error) {
	test_logs_directory, err := GetTestLogsDirectory()
	if err != nil {
		return "", err
	}

	profilerReportPath := filepath.Join(test_logs_directory, "test_profiler.txt")

	return profilerReportPath, nil
}
