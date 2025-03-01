package sc2_map_processing

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/chunk_utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	"github.com/icza/mpq"
	"github.com/icza/s2prot/rep"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// ReplayProcessingChannelContents is a struct that is used to pass data
// between the orchestrator and the workers in the pipeline.
type ReplayMapExtractProcessingChannel struct {
	Index        int
	ChunkOfFiles []string
}

type ReplayProcessingMapInfo struct {
	ReplayFilename      string
	MapHashAndExtension string
}

type ExtractMapChannelContents struct {
	mapOfURLs map[url.URL]ReplayProcessingMapInfo
}

// GetAllReplaysMapURLs retrieves the map URLs from the replay files.
func GetAllReplaysMapURLs(
	files []string,
	mapsOnDriveSet map[string]struct{},
	cliFlags utils.CLIFlags,
) (
	map[url.URL]string,
	error,
) {

	// Setting up the progress bar to handle the processing:
	nFiles := len(files)
	progressBarLen := nFiles
	progressBar := utils.NewProgressBar(
		progressBarLen,
		"[1/4] Retrieving all map URLs: ",
	)
	defer progressBar.Close()

	// If it is specified by the user to perform the processing without
	// multiprocessing GOMAXPROCS needs to be set to 1 in order to allow 1 thread:
	runtime.GOMAXPROCS(cliFlags.NumberOfThreads)
	inputChannel := make(chan ReplayMapExtractProcessingChannel, cliFlags.NumberOfThreads+1)
	outputChannel := make(chan ExtractMapChannelContents, cliFlags.NumberOfThreads+1)

	// Creating chunks of files for data parallel multiprocessing:
	downloadMapsChunks, _ := chunk_utils.GetChunkListAndPackageBool(
		files,
		0,
		cliFlags.NumberOfThreads,
		len(files),
	)

	// Initializing the wait group with the selected number of workers/threads:
	var wg sync.WaitGroup
	// Adding a task for each of the supplied chunks to speed up the processing:
	wg.Add(cliFlags.NumberOfThreads)

	// Spin up workers waiting for chunks to process:
	for i := 0; i < cliFlags.NumberOfThreads; i++ {
		go createMapExtractingGoroutines(
			inputChannel,
			outputChannel,
			progressBar,
			&wg,
		)
	}

	// Handling the closing of the channels and waiting for the workers to finish.
	// We are using channels to alleviate the issue with mutexes.
	// After all of the data comes through the output channels,
	// the information about which maps should be downloaded will be put into a single
	// map which will handle the duplicates:
	// Passing the chunks to input channel before the workers start processing:
	for index, chunk := range downloadMapsChunks {
		inputChannel <- ReplayMapExtractProcessingChannel{
			Index:        index,
			ChunkOfFiles: chunk,
		}
	}

	// No more data will be sent to the workers:
	close(inputChannel)
	go func() {
		// No more chunks will be sent to the workers,
		// they are all consumed above:
		wg.Wait()
		// No more output will come after the wait group is done:
		close(outputChannel)
	}()

	// Consume the output from the workers. This is needed to get rid of the
	// duplicate map URLs, multiple replays can have the same map:
	toDownloadURLToFileMap := make(map[url.URL]string)
	for output := range outputChannel {
		for url := range output.mapOfURLs {

			replayHashExtension := output.mapOfURLs[url].MapHashAndExtension

			// The map will have to be downloaded only if it is not already existing
			// on the disk:
			_, ok := mapsOnDriveSet[replayHashExtension]
			if ok {
				// the map is already downloaded, skip it:
				log.WithField("map", replayHashExtension).
					Debug("Map is already downloaded, continuing.")
				continue
			}

			toDownloadURLToFileMap[url] = replayHashExtension
		}
	}

	log.WithField("nMapsToDownload", len(toDownloadURLToFileMap)).
		Debug("Finished GetAllReplaysMapURLs()")

	// Return all of the URLs
	return toDownloadURLToFileMap, nil
}

// createMapExtractingGoroutines creates the goroutines that process the replay files
// and extract the map URLs.
func createMapExtractingGoroutines(
	inputChannel chan ReplayMapExtractProcessingChannel,
	outputChannel chan ExtractMapChannelContents,
	progressBar *progressbar.ProgressBar,
	wg *sync.WaitGroup,
) {

	// Running the goroutine until the input channel is closed:
	for {
		channelContents, ok := <-inputChannel
		if !ok {
			wg.Done()
			return
		}
		// Process the chunk of files and add the URLs to the map
		mapOfURLs := make(map[url.URL]ReplayProcessingMapInfo)
		for _, replayFullFilepath := range channelContents.ChunkOfFiles {
			// Filling out the map of urls that will be returned through the output channel
			// the caller will handle the consumption of the output channel and
			// therefore the deduplication of map URLs.
			// Error handling is done inside the function, if the processing fails,
			// the function returns false, error is logged, and the processing continues:
			if !processFileExtractMapURL(
				progressBar,
				replayFullFilepath,
				mapOfURLs,
			) {
				continue
			}
		}
		// send to output channel:
		log.WithField("len_mapOfURLS", len(mapOfURLs)).
			Info("Sending to output channel")

		outputChannel <- ExtractMapChannelContents{
			mapOfURLs: mapOfURLs,
		}
	}

}

// processFileExtractMapURL processes the replay file to extract the map URL and hash.
func processFileExtractMapURL(
	progressBar *progressbar.ProgressBar,
	replayFullFilepath string,
	urls map[url.URL]ReplayProcessingMapInfo,
) bool {

	// Lambda to process the replay file to have
	// deferred progress bar increment:

	// Defer the progress bar increment:
	defer func() {
		if err := progressBar.Add(1); err != nil {
			log.WithField("error", err).
				Error("Error updating progress bar in GetAllReplaysMapURLs")
		}
	}()
	replayFilename := filepath.Base(replayFullFilepath)

	mapURL, mapHashAndExtension, err := getURL(replayFullFilepath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"replayFile": replayFullFilepath,
		}).Error("Failed to get map URL from replay")
		return false
	}

	// Store the map URL and hash:
	urls[mapURL] = ReplayProcessingMapInfo{
		ReplayFilename:      replayFilename,
		MapHashAndExtension: mapHashAndExtension,
	}

	return true
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

// GetMapURLAndHashFromReplayData extracts the map URL,
// hash, and file extension from the replay data.
func GetMapURLAndHashFromReplayData(
	replayData *rep.Rep,
) (url.URL, string, bool) {
	log.Debug("Entered getMapURLAndHashFromReplayData()")
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
	log.Debug("Finished getMapURLAndHashFromReplayData()")
	return *mapURL, hashAndExtensionMerged, true
}

// ReadLocalizedDataFromMapGetForeignToEnglishMapping opens the map file (MPQ),
// reads the listfile, finds the english locale file,
// reads the map name and returns it.
func ReadLocalizedDataFromMapGetForeignToEnglishMapping(
	mapFilepath string,
	progressBar *progressbar.ProgressBar,
) (map[string]string, error) {
	log.Debug("Entered readLocalizedDataFromMap()")

	defer func() {
		err := progressBar.Add(1)
		if err != nil {
			log.WithField("error", err).
				Error("Error updating progress bar in readLocalizedDataFromMap")
		}
	}()

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

	log.Debug("Finished readLocalizedDataFromMap()")
	return foreignToEnglishMapName, nil
}

// findEnglishLocaleFile looks for the file containing the english map name
func findLocaleFiles(MPQArchiveBytes []byte) ([]string, string, error) {
	log.Debug("Entered findEnglishLocaleFile()")

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
		if strings.
			Contains(fileNameString, "SC2Data\\LocalizedData\\GameStrings") {
			localizationFiles = append(localizationFiles, fileNameString)
			foundLocaleFile = true
		}
		// Only English locale file:
		if strings.
			HasPrefix(fileNameString, "enUS.SC2Data\\LocalizedData\\GameStrings") {
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

	log.Debug("Finished findEnglishLocaleFile()")
	return localizationFiles, englishLocaleFile, nil
}

// readLocaleFileGetMapName reads the map name from the locale file within the MPQ archive.
func readLocaleFileGetMapName(
	mpqArchive *mpq.MPQ,
	localeFileName string,
) (string, error) {

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
	log.Debug("Entered cleanMapName()")

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

	log.Debug("Finished cleanMapName()")
	return mapName
}

// getMapNameFromLocaleFile reads the english map name
// from the bytes of opened locale file.
func getMapNameFromLocaleFile(MPQLocaleFileBytes []byte) (string, error) {
	log.Debug("Entered getMapNameFromLocaleFile()")

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

	log.Debug("Finished getMapNameFromLocaleFile(), found map name.")
	return mapName, nil
}

// replaceNewlinesSplitData replaces the line endings "\r\n" with "\n".
func replaceNewlinesSplitData(input string) []string {
	replacedNewlines := strings.ReplaceAll(input, "\r\n", "\n")
	splitFile := strings.Split(replacedNewlines, "\n")

	return splitFile
}

// SaveForeignToEnglishMappingToDrive saves the foreign to english mapping to the drive.
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
