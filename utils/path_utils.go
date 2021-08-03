package utils

import (
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func ListFiles(inputPath string, fileExtension string) []string {

	log.Info("Entered ListFiles()")

	files, err := ioutil.ReadDir(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var listOfReplayFiles []string
	for _, file := range files {
		filename := file.Name()
		fileExtension := filepath.Ext(filename)
		if fileExtension != fileExtension {
		} else {
			absoluteReplayPath := filepath.Join(inputPath, filename)
			listOfReplayFiles = append(listOfReplayFiles, absoluteReplayPath)
		}
	}

	log.Info("Finished ListFiles()")

	return listOfReplayFiles

}
