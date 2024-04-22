package dataproc

import (
	"fmt"
	"strings"

	"github.com/icza/mpq"
	log "github.com/sirupsen/logrus"
)

func readLocalizedDataFromMap(mapFilepath string) (string, error) {
	log.Info("Entered readLocalizedDataFromMap()")

	m, err := mpq.NewFromFile(mapFilepath)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "err": err}).
			Error("Finished readLocalizedDataFromMap(), Error reading map file with MPQ: ")
		return "", err
	}
	defer m.Close()

	data, err := m.FileByName("(listfile)")
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "err": err}).
			Error("Finished readLocalizedDataFromMap() Error reading listfile from MPQ: ")
		return "", err
	}

	localizationMPQFileName, err := findEnglishLocaleFile(data)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "err": err}).
			Error("Finished readLocalizedDataFromMap() Error finding english locale file: ")
		return "", err
	}

	localeFileDataBytes, err := m.FileByName(localizationMPQFileName)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "err": err}).
			Error("Finished readLocalizedDataFromMap() Error reading locale file from MPQ: ")
		return "", err
	}

	mapName, err := getMapNameFromLocaleFile(localeFileDataBytes)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "err": err}).
			Error("Finished readLocalizedDataFromMap() Error getting map name from locale file: ")
		return "", err
	}

	log.Info("Finished readLocalizedDataFromMap()")
	return mapName, nil
}

func findEnglishLocaleFile(MPQArchiveBytes []byte) (string, error) {
	log.Info("Entered findEnglishLocaleFile()")

	// Cast bytes to string:
	MPQStringData := string(MPQArchiveBytes)
	// Split data by newline:
	splitListfile := replaceNewlinesSplitData(MPQStringData)
	// Look for the file containing the map name:
	foundLocaleFile := false
	localizationMPQFileName := ""
	log.WithField("files", splitListfile).Debug("List of files inside archive")
	for _, fileNameString := range splitListfile {
		if strings.HasPrefix(fileNameString, "enUS.SC2Data\\LocalizedData\\GameStrings") {
			foundLocaleFile = true
			localizationMPQFileName = fileNameString
			break
		}
	}
	if !foundLocaleFile {
		log.Error("Failed in findEnglishLocaleFile()")
		return "", fmt.Errorf("could not find localization file in MPQ")
	}

	log.Info("Finished findEnglishLocaleFile()")
	return localizationMPQFileName, nil
}

func getMapNameFromLocaleFile(MPQLocaleFileBytes []byte) (string, error) {

	log.Info("Entered getMapNameFromLocaleFile()")

	// Cast File content into string:
	localeFileDataString := string(MPQLocaleFileBytes)
	splitLocaleFileString := replaceNewlinesSplitData(localeFileDataString)
	// Look for field with the map name:
	mapNameFound := false
	mapName := ""
	fieldPrefix := "DocInfo/Name="
	for _, field := range splitLocaleFileString {
		if strings.HasPrefix(field, fieldPrefix) {
			mapNameFound = true
			mapName = strings.TrimPrefix(field, fieldPrefix)
			break
		}
	}
	if !mapNameFound {
		log.Error("Failed in getMapNameFromLocaleFile()")
		return "", fmt.Errorf("map name was not found")
	}

	log.Info("Finished getMapNameFromLocaleFile(), found map name.")
	return mapName, nil
}
