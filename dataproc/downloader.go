package dataproc

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/alitto/pond"
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
	mapDownloadDirectory string                                           // NOT_MODIFIABLE Directory where the maps are downloaded
	existingMapFiles     *map[string]struct{}                             // NOT_MODIFIABLE List of existing map files in the maps directory
	mapHashAndTypeToName *map[string]string                               // MODIFIABLE Mapping from filename to english map name
	currentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo // MODIFIABLE Mapping from filename to list of channels to be notified when download finishes
	sharedRWMutex        *sync.RWMutex                                    // MODIFIABLE Mutex for shared state
	workerPool           *pond.WorkerPool                                 // Worker pool for downloading maps.
}

// Constructor for new downloader shared state
// NewDownloaderSharedState creates a new DownloaderSharedState.
func NewDownloaderSharedState(
	existingMapFilesList []string,
	maxPoolCapacity int) DownloaderSharedState {

	// TODO: Change from a slice of strings into a set implementation
	mapFilesSetToPopulate := make(map[string]struct{})
	for _, existingMap := range existingMapFilesList {
		_, exists := mapFilesSetToPopulate[existingMap]
		if !exists {
			mapFilesSetToPopulate[existingMap] = struct{}{}
		}
	}

	return DownloaderSharedState{
		mapDownloadDirectory: "maps",
		existingMapFiles:     &mapFilesSetToPopulate,
		mapHashAndTypeToName: &map[string]string{},
		currentlyDownloading: &map[string][]chan DownloadTaskReturnChannelInfo{},
		sharedRWMutex:        &sync.RWMutex{},
		workerPool:           pond.New(3, maxPoolCapacity, pond.Strategy(pond.Eager())),
	}
}

// DownloadTaskState holds all of the information needed for the download task.
type DownloadTaskState struct {
	mapDownloadDirectory string
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

	// TODO: Read and verify if the map exists here it doesn't require locking:
	maybeMapName, err := getMapNameFromDrive(*downloaderSharedState, mapHashAndType)
	if maybeMapName == "" && err != nil {
		return ""
	}

	englishMapName, downloadTaskInfoChannel := dispatchDownloadTask(*downloaderSharedState, mapHashAndType, mapURL)

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

func getMapNameFromDrive(
	downloaderSharedState DownloaderSharedState,
	mapHashAndType string) (string, error) {

	// Check if the map name is already downloaded:

	// TODO: read the english map name from map file
	mapPath := path.Join(downloaderSharedState.mapDownloadDirectory, mapHashAndType)
	englishMapName, err := readLocalizedDataFromMap(mapPath)
	if err != nil {
		return "", err
	}

	if englishMapName == "" {
		return "", fmt.Errorf("map name was read but it is empty")
	}

	// Locking access to shared state:
	downloaderSharedState.sharedRWMutex.Lock()
	defer downloaderSharedState.sharedRWMutex.Unlock()

	// REVIEW: Verify this:
	// Add to the mapHashAndTypeToName if the map is on drive:
	// Add the variable to the mapHashAndTypeToName
	// within the mutex lock to avoid IO operations while under lock.
	(*downloaderSharedState.mapHashAndTypeToName)[mapHashAndType] = englishMapName

	return englishMapName, nil
}

// TODO: Early return and log errors, create channel with error present.
// log process ID, thread ID, log when new download task comes
func downloadSingleMap(taskState DownloadTaskState) {

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

func dispatchDownloadTask(
	downloaderSharedState DownloaderSharedState,
	mapHashAndType string,
	mapURL url.URL) (string, chan DownloadTaskReturnChannelInfo) {

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
				downloadSingleMap(taskState)
			},
		)
	}
	return "", downloadTaskInfoChannel

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
