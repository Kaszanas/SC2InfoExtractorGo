package test_utils

import (
	"archive/zip"

	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

func TestFileSetup() error {
	testFilesDirectory, err := settings.GetTestFilesDirectory()
	if err != nil {
		return err
	}

	testReplaysDirectory, err := settings.GetTestInputDirectory()
	if err != nil {
		return err
	}

	downloadFilepath := filepath.Join(testFilesDirectory, settings.TEST_ARCHIVE_FILEPATH)

	archive, err := DownloadFile(downloadFilepath, settings.TEST_FILES_ARCHIVE)
	if err != nil {
		return err
	}

	// Unpack the archive, keeping only the files that do not already exist.
	err = UnpackArchive(archive.Name(), testReplaysDirectory)
	if err != nil {
		return err
	}

	return nil
}

func UnpackArchive(archivePath string, destinationDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, archivedFile := range reader.File {
		pathToExtractedFile := filepath.Join(destinationDir, archivedFile.Name)

		// Skip if the file already exists:
		if _, err := os.Stat(pathToExtractedFile); !os.IsNotExist(err) {
			continue
		}

		// Create file and copy contents
		err := os.MkdirAll(filepath.Dir(pathToExtractedFile), os.ModePerm)
		if err != nil {
			return err
		}

		extractedFile, err := os.OpenFile(pathToExtractedFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, archivedFile.Mode())
		if err != nil {
			return err
		}
		defer extractedFile.Close()

		archivedFileReadCloser, err := archivedFile.Open()
		if err != nil {
			return err
		}
		defer archivedFileReadCloser.Close()

		// Copy the file to the destination outside of the archive:
		_, err = io.Copy(extractedFile, archivedFileReadCloser)
		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadFile sends a GET request to the url and writes the response body to the filepath.
func DownloadFile(downloadFilepath string, url string) (*os.File, error) {

	// Get the data
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// TODO: Check if the response is of file type.
	// Create the file
	outputFile, err := os.Create(downloadFilepath)
	if err != nil {
		return nil, err
	}
	defer outputFile.Close()

	// Write the body to file
	_, err = io.Copy(outputFile, response.Body)
	return outputFile, err

}
