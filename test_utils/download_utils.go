package test_utils

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

// TODO: This is only a sample, the repository name will change:
var TEST_FILES_ARCHIVE = "https://github.com/Kaszanas/SC2InfoExtractorGo/releases/latest/download/sc2reset_test_files.zip"
var TEST_ARCHIVE_FILEPATH = "test_files.zip"

func TestFileSetup() error {

	testFilesDirectory, err := settings.GetTestFilesDirectory()
	if err != nil {
		return err
	}

	archive, err := DownloadTestFiles(TEST_FILES_ARCHIVE, testFilesDirectory)
	if err != nil {
		return err
	}

	// TODO: Unpack the archive, keeping only the files that do not already exist.

}

func DownloadTestFiles(testFilesURL string, downloadDir string) (*os.File, error) {

	testFilesDirectory, err := settings.GetTestFilesDirectory()
	if err != nil {
		return nil, err
	}

	downloadFilepath := filepath.Join(testFilesDirectory, TEST_ARCHIVE_FILEPATH)

	archive, err := DownloadFile(downloadFilepath, testFilesURL)
	if err != nil {
		return nil, err
	}

	return archive, nil

}

// DownloadFile sends a GET request to the url and writes the response body to the filepath.
func DownloadFile(downloadFilepath string, url string) (*os.File, error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: Check if the response is of file type.

	// Create the file
	out, err := os.Create(downloadFilepath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return out, err

}
