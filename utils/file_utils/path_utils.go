package file_utils

import (
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ListFiles creates a slice of filepaths from a give input directory
// based filtering supplied fileExtension
func ListFiles(
	inputPath string,
	filterFileExtension string,
) ([]string, error) {

	log.WithFields(log.Fields{
		"inputPath":           inputPath,
		"filterFileExtension": filterFileExtension}).
		Info("Entered ListFiles()")

	files, err := os.ReadDir(inputPath)
	if err != nil {
		log.WithField("error", err).Error("Error reading directory")
		return nil, err
	}

	var listOfFiles []string
	if filterFileExtension == "" {
		listOfFiles = getAllFiles(files, inputPath)
		return listOfFiles, nil
	}

	listOfFiles = getFilesByExtension(files, inputPath, filterFileExtension)

	log.WithField("n_files", len(listOfFiles)).Info("Finished ListFiles()")
	return listOfFiles, nil
}

// ExistingFilesSet creates a set of existing files in a directory.
func ExistingFilesSet(
	inputPath string,
	fiterFileExtension string,
) (map[string]struct{}, error) {

	log.WithFields(log.Fields{
		"inputPath":           inputPath,
		"filterFileExtension": fiterFileExtension}).
		Info("Entered ExistingFilesSet()")

	// List the files in the selected directory:
	listOfFiles, err := ListFiles(inputPath, fiterFileExtension)
	if err != nil {
		log.WithField("error", err).Error("Error getting list of files")
		return nil, err
	}

	// Convert from a slice (list) to a set (map),
	// by default files should not be duplicates:
	existingFilesSet := make(map[string]struct{})
	for _, file := range listOfFiles {
		existingFilesSet[file] = struct{}{}
	}

	return existingFilesSet, nil
}

// getFilesByExtension filters files by extension and returns a slice of filepaths
func getFilesByExtension(
	files []fs.DirEntry,
	inputPath string,
	filterFileExtension string) []string {
	var listOfFiles []string
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			fileExtension := filepath.Ext(filename)
			if fileExtension == filterFileExtension {
				absoluteReplayPath := filepath.Join(inputPath, filename)
				listOfFiles = append(listOfFiles, absoluteReplayPath)
			}
		}
	}
	return listOfFiles
}

// getAllFiles returns a slice of filepaths for all files in a directory
func getAllFiles(files []fs.DirEntry, inputPath string) []string {
	var listOfFiles []string
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			absoluteReplayPath := filepath.Join(inputPath, filename)
			listOfFiles = append(listOfFiles, absoluteReplayPath)
		}
	}
	log.Info("Finished ListFiles()")
	return listOfFiles
}
