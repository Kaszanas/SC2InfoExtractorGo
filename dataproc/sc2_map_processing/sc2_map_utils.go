package sc2_map_processing

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/icza/mpq"
	"github.com/icza/s2prot/rep"
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
	processedReplaysFilepath string,
	mapsDirectory string,
	cliFlags utils.CLIFlags,
) (
	map[url.URL]string,
	persistent_data.ProcessedReplaysToFileInfo,
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

	processedReplays, err := persistent_data.OpenOrCreateProcessedReplaysToFileInfo(
		processedReplaysFilepath,
		mapsDirectory,
		fileChunks,
	)
	if err != nil {
		return nil, persistent_data.ProcessedReplaysToFileInfo{}, err
	}

	processedReplaysSyncMap := processedReplays.ConvertToSyncMap()

	// Spin up workers waiting for chunks to process:
	for i := 0; i < cliFlags.NumberOfThreads; i++ {
		go func() {
			for {
				channelContents, ok := <-channel
				if !ok {
					wg.Done()
					return
				}
				// Process the chunk of files and add the URLs to the map
				for _, replayFullFilepath := range channelContents.ChunkOfFiles {

					replayFilename := filepath.Base(replayFullFilepath)

					// Check if the replay was already processed:
					fileInfo, err := os.Stat(replayFullFilepath)
					if err != nil {
						log.WithFields(log.Fields{
							"error":      err,
							"replayFile": replayFullFilepath,
						}).Error("Failed to get file info.")
						continue
					}
					fileInfoToCheck, alreadyProcessed :=
						processedReplaysSyncMap.Load(replayFilename)
					if alreadyProcessed {
						// Check if the file was modified since the last time it was processed:
						if persistent_data.CheckFileInfoEq(
							fileInfo,
							fileInfoToCheck.(persistent_data.FileInformationToCheck),
						) {
							// It is the same so continue
							continue
						}
						// It wasn't the same so the replay should be processed again:
						processedReplaysSyncMap.Delete(replayFilename)
					}

					// Assume getURLsFromReplay is a function that
					// returns a slice of URLs from a replay file
					replayData, err := rep.NewFromFile(replayFullFilepath)
					if err != nil {
						log.WithFields(log.Fields{"file": replayFullFilepath, "error": err}).
							Error("Failed to read replay file to retrieve map data")
						continue
					}

					mapURL, mapHashAndExtension, mapRetrieved :=
						GetMapURLAndHashFromReplayData(replayData)
					if !mapRetrieved {
						log.WithField("file", replayFullFilepath).
							Error("Failed to get map URL and hash from replay data")
						continue
					}
					urls.Store(mapURL, mapHashAndExtension)
					processedReplaysSyncMap.Store(
						replayFilename,
						persistent_data.FileInformationToCheck{
							LastModified: fileInfo.ModTime().Unix(),
							Size:         fileInfo.Size(),
						},
					)
				}
			}
		}()
	}

	// Passing the chunks to the workers:
	for index, chunk := range fileChunks {
		channel <- ReplayMapProcessingChannelContents{
			Index:        index,
			ChunkOfFiles: chunk}
	}

	close(channel)
	wg.Wait()

	processedReplaysReturn := persistent_data.
		FromSyncMapToProcessedReplaysToFileInfo(processedReplaysSyncMap)

	urlMapToFilename := convertFromSyncMapToURLMap(urls)

	// Save the processed replays to the file:
	err = processedReplaysReturn.SaveProcessedReplaysFile(processedReplaysFilepath)
	if err != nil {
		log.WithField("processedReplaysFile", processedReplaysFilepath).
			Error("Failed to save the processed replays file.")
		return nil, persistent_data.ProcessedReplaysToFileInfo{}, err
	}

	// Return all of the URLs
	return urlMapToFilename, processedReplaysReturn, nil
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
