package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
)

func listReplayFiles(inputPath string) []string {

	files, err := ioutil.ReadDir(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var listOfReplayFiles []string
	for _, file := range files {
		filename := file.Name()
		fileExtension := filepath.Ext(filename)
		if fileExtension != ".SC2Replay" {
		} else {
			absoluteReplayPath := filepath.Join(inputPath, filename)
			listOfReplayFiles = append(listOfReplayFiles, absoluteReplayPath)
		}
	}

	return listOfReplayFiles

}
