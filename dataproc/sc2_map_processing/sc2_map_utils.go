package sc2_map_processing

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/icza/mpq"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// GetMapURLAndHashFromReplayData extracts the map URL,
// hash, and file extension from the replay data.
func GetMapURLAndHashFromReplayData(replayData *rep.Rep) (url.URL, string, bool) {
	log.Info("Entered getMapURLAndHashFromReplayData()")
	cacheHandles := replayData.Details.CacheHandles()

	// Get the cacheHandle for the map, I am not sure whi is it the last CacheHandle:
	mapCacheHandle := cacheHandles[len(cacheHandles)-1]
	region := mapCacheHandle.Region

	badRegions := []string{"Unknown", "Public Test"}
	for _, badRegion := range badRegions {
		if region.Name == badRegion {
			log.WithField("region", region.Name).Error("Detected bad region!")
			return url.URL{}, "", false
		}
	}

	// SEA Region was removed, so its depot url won't work, replacing with US:
	if region.Name == "SEA" {
		log.WithField("region", region.Name).
			Info("Detected SEA region, replacing with US")
		region = rep.RegionUS
	}

	depotURL := region.DepotURL

	hashAndExtensionMerged := fmt.Sprintf(
		"%s.%s",
		mapCacheHandle.Digest,
		mapCacheHandle.Type,
	)
	mapURL := depotURL.JoinPath(hashAndExtensionMerged)
	log.Info("Finished getMapURLAndHashFromReplayData()")
	return *mapURL, hashAndExtensionMerged, true
}

// ReadLocalizedDataFromMapGetForeignToEnglishMapping opens the map file (MPQ),
// reads the listfile, finds the english locale file,
// reads the map name and returns it.
func ReadLocalizedDataFromMapGetForeignToEnglishMapping(
	mapFilepath string,
) (map[string]string, error) {
	log.Info("Entered readLocalizedDataFromMap()")

	mpqArchive, err := mpq.NewFromFile(mapFilepath)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "error": err}).
			Error("Finished readLocalizedDataFromMap(), Error reading map file with MPQ: ")
		return nil, err
	}
	defer mpqArchive.Close()

	data, err := mpqArchive.FileByName("(listfile)")
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "error": err}).
			Error("Finished readLocalizedDataFromMap() Error reading listfile from MPQ: ")
		return nil, err
	}

	listOfLocaleFiles, englishLocaleFile, err := findLocaleFiles(data)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "error": err}).
			Error("Finished readLocalizedDataFromMap() Couldn't find locale files")
		return nil, fmt.Errorf("couldn't find locale files: %s", err)
	}

	// Find english map name first, this is used to create the mapping from
	// the foreign map name to the english map name.
	englishMapName, err := readLocaleFileGetMapName(mpqArchive, englishLocaleFile)
	if err != nil {
		log.WithFields(log.Fields{"mapFilepath": mapFilepath, "error": err}).
			Error("Finished readLocalizedDataFromMap() Couldn't find english map name")
		return nil, fmt.Errorf("couldn't find english map name: %s", err)
	}

	// Create the mapping from the foreign map name to the english map name:
	foreignToEnglishMapName := make(map[string]string)
	for _, localizationMPQFileName := range listOfLocaleFiles {
		mapName, err := readLocaleFileGetMapName(mpqArchive, localizationMPQFileName)
		if err != nil {
			log.WithFields(log.Fields{
				"mapFilepath":             mapFilepath,
				"error":                   err,
				"localizationMPQFileName": localizationMPQFileName,
			}).
				Error("Finished readLocalizedDataFromMap() Couldn't get one of the map names.")
			return nil, fmt.Errorf("couldn't find map name: %s", err)
		}
		foreignToEnglishMapName[mapName] = englishMapName
	}

	log.Info("Finished readLocalizedDataFromMap()")
	return foreignToEnglishMapName, nil
}

// findEnglishLocaleFile looks for the file containing the english map name
func findLocaleFiles(MPQArchiveBytes []byte) ([]string, string, error) {
	log.Info("Entered findEnglishLocaleFile()")

	// Cast bytes to string:
	MPQStringData := string(MPQArchiveBytes)
	// Split data by newline:
	splitListfile := replaceNewlinesSplitData(MPQStringData)
	// Look for the file containing the map name:
	foundLocaleFile := false
	log.WithField("files", splitListfile).Debug("List of files inside archive")
	var localizationFiles []string
	englishLocaleFile := ""
	for _, fileNameString := range splitListfile {
		// All locale files:
		if strings.Contains(fileNameString, "SC2Data\\LocalizedData\\GameStrings") {
			localizationFiles = append(localizationFiles, fileNameString)
			foundLocaleFile = true
		}
		// Only English locale file:
		if strings.HasPrefix(fileNameString, "enUS.SC2Data\\LocalizedData\\GameStrings") {
			englishLocaleFile = fileNameString
			foundLocaleFile = true
		}

	}
	if !foundLocaleFile {
		log.Error("Failed in findEnglishLocaleFile()")
		return nil, "", fmt.Errorf("could not find any localization file in MPQ")
	}
	if englishLocaleFile == "" {
		log.Error("Failed in findEnglishLocaleFile()")
		return nil, "", fmt.Errorf("could not find english localization file in MPQ")
	}

	log.Info("Finished findEnglishLocaleFile()")
	return localizationFiles, englishLocaleFile, nil
}

func readLocaleFileGetMapName(mpqArchive *mpq.MPQ, localeFileName string) (string, error) {

	localeFileDataBytes, err := mpqArchive.FileByName(localeFileName)
	if err != nil {
		log.WithFields(log.Fields{"localeFileName": localeFileName, "err": err}).
			Error("Finished readLocaleFileGetMapName() Error reading locale file from MPQ: ")
		return "", err
	}

	mapName, err := getMapNameFromLocaleFile(localeFileDataBytes)
	if err != nil {
		log.WithFields(log.Fields{"localeFileName": localeFileName, "err": err}).
			Error("Finished readLocaleFileGetMapName() Error getting map name from locale file: ")
		return "", err
	}

	return mapName, nil
}

// getMapNameFromLocaleFile reads the english map name
// from the bytes of opened locale file.
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

func replaceNewlinesSplitData(input string) []string {
	replacedNewlines := strings.ReplaceAll(input, "\r\n", "\n")
	splitFile := strings.Split(replacedNewlines, "\n")

	return splitFile
}
