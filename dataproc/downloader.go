package dataproc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/alitto/pond"
	"github.com/icza/mpq"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// Mapping from hash and type to the name of the map
// Set of maps that are currently being downloaded,
// this avoids starting multiple downloads of the same map:
// REVIEW: How to effectively wait for the download to finish
// REVIEW: if another thread requests the same map name?
// channel of string because the response is the map name:
// DownloaderSharedState holds all of the shared state for the downloader.
type DownloaderSharedState struct {
	mapDownloadDirectory string                                           // Directory where the maps are downloaded
	existingMapFiles     *[]string                                        // List of existing map files in the maps directory
	mapHashAndTypeToName *map[string]string                               // Mapping from filename to english map name
	currentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo // Mapping from filename to list of channels to be notified when download finishes
	sharedRWMutex        *sync.RWMutex                                    // Mutex for shared state
	workerPool           *pond.WorkerPool                                 // Worker pool for downloading maps.
}

// Constructor for new downloader shared state
// NewDownloaderSharedState creates a new DownloaderSharedState.
func NewDownloaderSharedState(
	existingMapFiles []string,
	maxPoolCapacity int) DownloaderSharedState {

	return DownloaderSharedState{
		mapDownloadDirectory: "maps",
		existingMapFiles:     &[]string{},
		mapHashAndTypeToName: &map[string]string{},
		currentlyDownloading: &map[string][]chan DownloadTaskReturnChannelInfo{},
		sharedRWMutex:        &sync.RWMutex{},
		workerPool:           pond.New(3, maxPoolCapacity, pond.Strategy(pond.Eager())),
	}
}

// DownloadTaskState holds all of the information needed for the download task.
type DownloadTaskState struct {
	mapDownloadDirectory string
	existingMapFiles     *[]string
	mapHashAndTypeToName *map[string]string
	currentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo
	mapHashAndType       string
	mapURL               url.URL
	sharedRWMutex        *sync.RWMutex
}

// DownloadTaskReturnChannelInfo holds the information needed to return after the download finishes.
type DownloadTaskReturnChannelInfo struct {
	mapNameString string
	err           error
}

func getMapURLAndHashFromReplayData(replayData *rep.Rep) (url.URL, string, bool) {
	log.Info("Entered getMapURLAndHashFromReplayData()")
	cacheHandles := replayData.Details.CacheHandles()

	// Get the cacheHandle for the map, I am not sure whi is it the last CacheHandle:
	mapCacheHandle := cacheHandles[len(cacheHandles)-1]
	region := mapCacheHandle.Region

	// TODO: This is the only place where errors can be handled
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

	hashAndTypeMerged := fmt.Sprintf("%s.%s", mapCacheHandle.Digest, mapCacheHandle.Type)
	mapURL := depotURL.JoinPath(hashAndTypeMerged)
	log.Info("Finished getMapURLAndHashFromReplayData()")
	return *mapURL, hashAndTypeMerged, true
}

func getEnglishMapNameDownloadIfNotExists(
	downloaderSharedState *DownloaderSharedState,
	mapHashAndType string,
	mapURL url.URL) string {
	log.WithFields(
		log.Fields{
			"downloaderSharedState": downloaderSharedState,
			"mapHashAndType":        mapHashAndType,
			"mapURL":                mapURL.String(),
		},
	).Info("Entered getEnglishMapNameDownloadIfNotExists()")

	englishMapName, downloadTaskInfoChannel := func() (string, chan DownloadTaskReturnChannelInfo) {
		// Locking access to shared state:
		downloaderSharedState.sharedRWMutex.Lock()
		defer downloaderSharedState.sharedRWMutex.Unlock()

		// Check if the english map name was already read from the drive, return if present:
		englishMapName, ok := (*downloaderSharedState.mapHashAndTypeToName)[mapHashAndType]
		if ok {
			return englishMapName, nil
		}

		// Create channel
		// REVIEW: Verify if this can be a struct like that:
		downloadTaskInfoChannel := make(chan DownloadTaskReturnChannelInfo)

		// Check if key is in currently downloading:
		listOfChannels, ok := (*downloaderSharedState.currentlyDownloading)[mapHashAndType]
		if ok {
			// If it is downloading then add the channel to the list of channels waiting for result
			// Map is being downloaded, add it to the list of currently downloading maps:
			log.Info("Map is being downloaded, adding channel to receive the result.")
			(*downloaderSharedState.currentlyDownloading)[mapHashAndType] = append(listOfChannels, downloadTaskInfoChannel)
		} else {
			// TODO: Add logging
			taskState := DownloadTaskState{
				mapDownloadDirectory: downloaderSharedState.mapDownloadDirectory,
				existingMapFiles:     downloaderSharedState.existingMapFiles,
				mapHashAndTypeToName: downloaderSharedState.mapHashAndTypeToName,
				currentlyDownloading: downloaderSharedState.currentlyDownloading,
				mapHashAndType:       mapHashAndType,
				mapURL:               mapURL,
				sharedRWMutex:        downloaderSharedState.sharedRWMutex,
			}
			// if it is not then add key to the map and create one element slice with the channel
			// and submit the download task to the worker pool:
			(*downloaderSharedState.currentlyDownloading)[mapHashAndType] = []chan DownloadTaskReturnChannelInfo{downloadTaskInfoChannel}
			downloaderSharedState.workerPool.Submit(
				func() {
					// REVIEW: How to recover errors:
					downloadSingleMapOrRetrieveFromDrive(taskState)
				},
			)
		}
		return "", downloadTaskInfoChannel
	}()

	if englishMapName != "" {
		return englishMapName
	}

	// Wait for channel to finish downloading the map.
	taskStatus := <-downloadTaskInfoChannel
	if taskStatus.err != nil {
		log.WithField("err", taskStatus.err).Error("Error downloading map")
		return ""
	}

	// Mutex unlock
	log.Info("Finished getEnglishMapNameDownloadIfNotExists()")
	return taskStatus.mapNameString
}

// TODO: Early return and log errors, create channel with error present.
// log process ID, thread ID, log when new download task comes
func downloadSingleMapOrRetrieveFromDrive(taskState DownloadTaskState) {

	outputFilepath := path.Join(taskState.mapDownloadDirectory, taskState.mapHashAndType)

	response, err := http.Get(taskState.mapURL.String())
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(&taskState, "", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			"",
			fmt.Errorf("request returned code other than 200 OK"),
		)
		return
	}

	// Create output file:
	outFile, err := os.Create(outputFilepath)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(&taskState, "", err)
		return
	}
	defer outFile.Close()

	// Copy contents of response to the file:
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(&taskState, "", err)
		return
	}

	englishMapName, err := readLocalizedDataFromMap(outputFilepath)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			"",
			err,
		)
		return
	}

	sendDownloadTaskReturnInfoToChannels(&taskState, englishMapName, nil)
}

func sendDownloadTaskReturnInfoToChannels(
	taskState *DownloadTaskState,
	englishMapName string,
	err error) {
	taskState.sharedRWMutex.Lock()
	defer taskState.sharedRWMutex.Unlock()

	(*taskState.mapHashAndTypeToName)[taskState.mapHashAndType] = englishMapName
	for _, channel := range (*taskState.currentlyDownloading)[taskState.mapHashAndType] {
		channel <- DownloadTaskReturnChannelInfo{
			mapNameString: englishMapName,
			err:           err,
		}
	}
	delete(*taskState.currentlyDownloading, taskState.mapHashAndType)
}

func replaceNewlinesSplitData(input string) []string {
	replacedNewlines := strings.ReplaceAll(input, "\r\n", "\n")
	splitFile := strings.Split(replacedNewlines, "\n")

	return splitFile
}

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
	fmt.Println("Files inside archive:", splitListfile)
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
