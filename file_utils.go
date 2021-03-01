package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	log "github.com/sirupsen/logrus"
)

func createProcessingInfoFile() (*os.File, data.ProcessingInfo) {
	processingInfoFile, err := os.OpenFile("processing.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to create or open the processing.log: ", err)
	}
	byteValue, err := ioutil.ReadAll(processingInfoFile)
	if err != nil {
		log.Fatal("Failed to read bytes from processing.log: ", err)
	}

	// This will hold: {"anonymizedPlayers": {"toon": id}, "packageCounter": int, "processedFiles": [path, path, path]}
	var processingInfoStruct data.ProcessingInfo
	err = json.Unmarshal(byteValue, &processingInfoStruct)
	if err != nil {
		processingInfoStruct = data.DefaultProcessingInfo()
		log.Error("Failed to unmarshall the processing.log, initializing empty data.ProcessingInfo struct")
	}

	return processingInfoFile, processingInfoStruct
}

func saveProcessingInfo(processingInfoFile os.File, processingInfoStruct data.ProcessingInfo) {

	processingInfoBytes, err := json.Marshal(processingInfoStruct)
	if err != nil {
		log.Fatal("Failed to marshal processingInfo that is used to create processing.log: ", err)
	}
	_, err = processingInfoFile.Write(processingInfoBytes)
	if err != nil {
		log.Fatal("Failed to save the processingInfoFile: ", err)
	}
}
