package test_utils

import (
	"archive/zip"

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

	testReplaysDirectory, err := settings.GetTestInputDirectory()
	if err != nil {
		return err
	}

	archive, err := DownloadTestFiles(TEST_FILES_ARCHIVE, testFilesDirectory)
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

	for _, file := range reader.File {
		filePath := filepath.Join(destinationDir, file.Name)

		// Skip if the file already exists:
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			continue
		}

		// Create file and copy contents
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		// Get the archived file:
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// Copy the file to the destination outside of the archive:
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
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
