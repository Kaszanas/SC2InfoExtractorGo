package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	log "github.com/sirupsen/logrus"
)

func createFile(filePath string) (os.File, []byte) {
	createdOrReadFile, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to create or open the processing.log: ", err)
		os.Exit(1)
	}
	byteValue, err := ioutil.ReadAll(createdOrReadFile)
	if err != nil {
		log.Fatal("Failed to read bytes from processing.log: ", err)
		os.Exit(1)
	}

	return *createdOrReadFile, byteValue
}

func createProcessingInfoFile() (*os.File, data.ProcessingInfo) {

	processingInfoFile, byteValue := createFile("processing.log")

	// This will hold: {"anonymizedPlayers": {"toon": id}, "packageCounter": int, "processedFiles": [path, path, path]}
	var processingInfoStruct data.ProcessingInfo
	err := json.Unmarshal(byteValue, &processingInfoStruct)
	if err != nil {
		processingInfoStruct = data.DefaultProcessingInfo()
		log.Error("Failed to unmarshall the processing.log, initializing empty data.ProcessingInfo struct")
	}

	return &processingInfoFile, processingInfoStruct
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

func unmarshalLocaleMapping(pathToMappingFile string) map[string]interface{} {
	localizedMapping := make(map[string]interface{})

	if !unmarshalFile(pathToMappingFile, &localizedMapping) {
		log.WithField("pathToMappingFile", pathToMappingFile).Error("Failed to open and unmarshal the mapping file!")
		return localizedMapping
	}

	return localizedMapping
}

func unmarshalPersistentAnonymizedNicknames(pathToMappingFile string) map[string]interface{} {
	persistentPlayerMapping := make(map[string]interface{})

	if !unmarshalFile(pathToMappingFile, &persistentPlayerMapping) {
		log.WithField("pathToMappingFile", pathToMappingFile).Error("Failed to open and unmarshal the mapping file!")
		return persistentPlayerMapping
	}

	return persistentPlayerMapping
}

func unmarshalFile(pathToMappingFile string, mappingToPopulate *map[string]interface{}) bool {
	var file, err = os.Open(pathToMappingFile)
	if err != nil {
		log.WithField("fileError", err.Error()).Info("Failed to open Localization Mapping file.")
		return false
	}
	defer file.Close()

	jsonBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.WithField("readError", err.Error()).Info("Failed to read Localization Mapping file.")
		return false
	}

	err = json.Unmarshal([]byte(jsonBytes), &mappingToPopulate)
	if err != nil {
		log.WithField("jsonMarshalError", err.Error()).Info("Could not unmarshal the Localization JSON file.")
	}

	return true
}
