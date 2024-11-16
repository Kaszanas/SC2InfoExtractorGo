package sc2_map_processing

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	"github.com/icza/mpq"
	"github.com/icza/s2prot/rep"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// ReplayProcessingChannelContents is a struct that is used to pass data
// between the orchestrator and the workers in the pipeline.
type ReplayMapProcessingChannelContents struct {
	Index        int
	ChunkOfFiles []string
}

func GetAllReplaysMapURLs(
	fileChunks [][]string,
	downloadedMapsForReplaysFilepath string,
	mapsDirectory string,
	cliFlags utils.CLIFlags,
) (
	map[url.URL]string,
	persistent_data.DownloadedMapsReplaysToFileInfo,
	error,
) {

	// If it is specified by the user to perform the processing without
	// multiprocessing GOMAXPROCS needs to be set to 1 in order to allow 1 thread:
	runtime.GOMAXPROCS(cliFlags.NumberOfThreads)
	var channel = make(chan ReplayMapProcessingChannelContents, cliFlags.NumberOfThreads+1)
	var wg sync.WaitGroup
	// Adding a task for each of the supplied chunks to speed up the processing:
	wg.Add(cliFlags.NumberOfThreads)

	// Create a sync.Map to store the URLs
	urls := &sync.Map{}

	downloadedMapsForReplays, err := persistent_data.
		OpenOrCreateDownloadedMapsForReplaysToFileInfo(
			downloadedMapsForReplaysFilepath,
			mapsDirectory,
			fileChunks,
		)
	if err != nil {
		return nil, persistent_data.DownloadedMapsReplaysToFileInfo{}, err
	}

	downloadedMapsForReplaysSyncMap := downloadedMapsForReplays.
		ConvertToSyncMap()

	// Progress bar logic:
	nChunks := len(fileChunks)
	nFiles := 0
	for _, chunk := range fileChunks {
		nFiles += len(chunk)
	}
	progressBarLen := nChunks * nFiles
	progressBar := utils.NewProgressBar(
		progressBarLen,
		"[1/4] Retrieving all map URLs: ",
	)
	defer progressBar.Close()
	// Spin up workers waiting for chunks to process:
	for i := 0; i < cliFlags.NumberOfThreads; i++ {
		go createMapExtractingGoroutines(
			channel,
			progressBar,
			urls,
			downloadedMapsForReplaysSyncMap,
			&wg,
		)
	}

	// Passing the chunks to the workers:
	for index, chunk := range fileChunks {
		channel <- ReplayMapProcessingChannelContents{
			Index:        index,
			ChunkOfFiles: chunk}
	}

	close(channel)
	wg.Wait()
	progressBar.Close()

	downloadedMapsForReplaysReturn := persistent_data.
		FromSyncMapToDownloadedMapsForReplaysToFileInfo(
			downloadedMapsForReplaysSyncMap,
		)

	urlMapToFilename := convertFromSyncMapToURLMap(urls)

	// Return all of the URLs
	return urlMapToFilename, downloadedMapsForReplaysReturn, nil
}

func createMapExtractingGoroutines(
	channel chan ReplayMapProcessingChannelContents,
	progressBar *progressbar.ProgressBar,
	urls *sync.Map,
	downloadedMapsForReplaysSyncMap *sync.Map,
	wg *sync.WaitGroup,
) {

	for {
		channelContents, ok := <-channel
		if !ok {
			wg.Done()
			return
		}
		// Process the chunk of files and add the URLs to the map
		for _, replayFullFilepath := range channelContents.ChunkOfFiles {

			processFileExtractMap(
				progressBar,
				replayFullFilepath,
				urls,
				downloadedMapsForReplaysSyncMap,
			)

		}
	}

}

func processFileExtractMap(
	progressBar *progressbar.ProgressBar,
	replayFullFilepath string,
	urls *sync.Map,
	downloadedMapsForReplaysSyncMap *sync.Map,
) {

	// Lambda to process the replay file to have
	// deferred progress bar increment:
	func() {
		// Defer the progress bar increment:
		defer func() {
			if err := progressBar.Add(1); err != nil {
				log.WithField("error", err).
					Error("Error updating progress bar in GetAllReplaysMapURLs")
			}
		}()
		replayFilename := filepath.Base(replayFullFilepath)

		// Check if the replay was already processed:
		fileInfo, err := os.Stat(replayFullFilepath)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"replayFile": replayFullFilepath,
			}).Error("Failed to get file info.")
			return
		}

		// If the file was already processed, and was not modified since,
		// then it will be skipped from getting the map URL:
		fileInfoToCheck, alreadyProcessed :=
			downloadedMapsForReplaysSyncMap.Load(replayFilename)
		if alreadyProcessed {
			// Check if the file was modified since the last time it was processed:
			if persistent_data.CheckFileInfoEq(
				fileInfo,
				fileInfoToCheck.(persistent_data.FileInformationToCheck),
			) {
				// It is the same so continue
				log.WithField("file", replayFullFilepath).
					Warning("This replay was already processed, map should be available, continuing!")
				return
			}
			// It wasn't the same so the replay should be processed again:
			log.WithField("file", replayFullFilepath).
				Warn("Replay was modified since the last time it was processed! Processing again.")
			downloadedMapsForReplaysSyncMap.Delete(replayFilename)
		}

		mapURL, mapHashAndExtension, err := getURL(replayFullFilepath)
		if err != nil {
			log.WithFields(log.Fields{
				"error":      err,
				"replayFile": replayFullFilepath,
			}).Error("Failed to get map URL from replay")
			return
		}

		urls.Store(mapURL, mapHashAndExtension)
		downloadedMapsForReplaysSyncMap.Store(
			replayFilename,
			persistent_data.FileInformationToCheck{
				LastModified: fileInfo.ModTime().Unix(),
				Size:         fileInfo.Size(),
			},
		)
	}()

}

// getURL retrieves the map URL from the replay file.
func getURL(replayFullFilepath string) (url.URL, string, error) {
	// Assume getURLsFromReplay is a function that
	// returns a slice of URLs from a replay file
	replayData, err := rep.NewFromFile(replayFullFilepath)
	if err != nil {
		log.WithFields(log.Fields{"file": replayFullFilepath, "error": err}).
			Error("Failed to read replay file to retrieve map data")
		return url.URL{}, "", err
	}

	mapURL, mapHashAndExtension, mapRetrieved :=
		GetMapURLAndHashFromReplayData(replayData)
	if !mapRetrieved {
		log.WithField("file", replayFullFilepath).
			Warning("Failed to get map URL and hash from replay data")
		return url.URL{}, "", fmt.Errorf("failed to get map URL and hash from replay data")
	}

	return mapURL, mapHashAndExtension, nil
}

// convertFromSyncMapToURLMap converts a sync.Map to a map[url.URL]string.
func convertFromSyncMapToURLMap(
	urls *sync.Map,
) map[url.URL]string {
	urlMap := make(map[url.URL]string)
	urls.Range(func(key, value interface{}) bool {
		urlMap[key.(url.URL)] = value.(string)
		return true
	})
	return urlMap
}

// GetMapURLAndHashFromReplayData extracts the map URL,
// hash, and file extension from the replay data.
func GetMapURLAndHashFromReplayData(
	replayData *rep.Rep,
) (url.URL, string, bool) {
	log.Info("Entered getMapURLAndHashFromReplayData()")
	cacheHandles := replayData.Details.CacheHandles()

	// Get the cacheHandle for the map, I am not sure whi is it the last CacheHandle:
	mapCacheHandle := cacheHandles[len(cacheHandles)-1]
	region := mapCacheHandle.Region

	unsupportedRegions := []string{"Unknown", "Public Test"}
	for _, badRegion := range unsupportedRegions {
		if region.Name == badRegion {
			log.WithField("region", region.Name).
				Warning(
					"Detected unsupported region! Won't download the map! Replay may fail further processing!",
				)
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
	progressBar *progressbar.ProgressBar,
) (map[string]string, error) {
	defer func() {
		if err := progressBar.Add(1); err != nil {
			log.WithField("error", err).
				Error("Error updating progress bar in ReadLocalizedDataFromMapGetForeignToEnglishMapping()")
		}
	}()
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
		// Get the foreign map name:
		mapName, err := readLocaleFileGetMapName(
			mpqArchive,
			localizationMPQFileName,
		)
		if err != nil {
			log.WithFields(log.Fields{
				"mapFilepath":             mapFilepath,
				"error":                   err,
				"localizationMPQFileName": localizationMPQFileName,
			}).
				Error("Finished readLocalizedDataFromMap() Couldn't get one of the map names.")
			return nil, fmt.Errorf("couldn't find map name: %s", err)
		}

		// https://github.com/Kaszanas/SC2InfoExtractorGo/issues/67
		// Clean the foreign map name:
		mapName = cleanMapName(mapName)

		foreignToEnglishMapName[mapName] = englishMapName
	}
	mpqArchive.Close()

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

// cleanMapName splits the map name by "\\\", and returns the first element
// (foreign map name) from the map name, otherwise the same map name will be returned.
func cleanMapName(mapName string) string {

	// Check if "\\\" exists in the map name, if it does
	// Check if the map name contains the substring "\\\"
	if !strings.Contains(mapName, "///") {
		return mapName
	}
	// Keep the left side of the string before "\\\".
	// Split the string by "\\\\" and keep only the first part
	mapName = strings.Split(mapName, "///")[0]
	// right trim the string to remove any trailing spaces
	mapName = strings.TrimRight(mapName, " ")

	return mapName
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

func SaveForeignToEnglishMappingToDrive(
	filepath string,
	foreignToEnglishMapping map[string]string,
) error {

	// Create or read the file:
	fileHandle, _, err := file_utils.ReadOrCreateFile(filepath)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to read or create the processed_replays.json file.")
		return err
	}

	// Save the mapping:
	mappingJSON, err := json.Marshal(foreignToEnglishMapping)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to marshal the foreignToEnglishMapping.")
		return err
	}

	_, err = fileHandle.Write(mappingJSON)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to write the foreignToEnglishMapping to the file.")
		return err
	}

	return nil
}
