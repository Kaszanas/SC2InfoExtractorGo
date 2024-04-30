package file_utils

import (
	"io/fs"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ListFiles creates a slice of filepaths from a give input directory
// based filtering supplied fileExtension
func ListFiles(inputPath string, filterFileExtension string) []string {

	log.WithFields(log.Fields{
		"inputPath":           inputPath,
		"filterFileExtension": filterFileExtension}).
		Info("Entered ListFiles()")

	files, err := os.ReadDir(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var listOfFiles []string
	if filterFileExtension == "" {
		listOfFiles = getAllFiles(files, inputPath)
		return listOfFiles
	}

	listOfFiles = getFilesByExtension(files, inputPath, filterFileExtension)

	log.WithField("n_files", len(listOfFiles)).Info("Finished ListFiles()")
	return listOfFiles
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
