package utils

import (
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// ListFiles creates a slice of filepaths from a give input directory based filtering supplied fileExtension
func ListFiles(inputPath string, filterFileExtension string) []string {

	log.Info("Entered ListFiles()")

	files, err := ioutil.ReadDir(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	if filterFileExtension == "" {
		var listOfReplayFiles []string
		for _, file := range files {
			filename := file.Name()
			absoluteReplayPath := filepath.Join(inputPath, filename)
			listOfReplayFiles = append(listOfReplayFiles, absoluteReplayPath)
		}
		log.Info("Finished ListFiles()")
		return listOfReplayFiles
	}

	var listOfReplayFiles []string
	for _, file := range files {
		filename := file.Name()
		fileExtension := filepath.Ext(filename)
		if fileExtension == filterFileExtension {
			absoluteReplayPath := filepath.Join(inputPath, filename)
			listOfReplayFiles = append(listOfReplayFiles, absoluteReplayPath)
		}
	}

	log.Info("Finished ListFiles()")
	return listOfReplayFiles
}
