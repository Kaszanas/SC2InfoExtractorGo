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
	mapHashAndTypeToName *map[string]string                               // MODIFIABLE Mapping from filename to english map name
	currentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo // MODIFIABLE Mapping from filename to list of channels to be notified when download finishes
	sharedRWMutex        *sync.RWMutex                                    // MODIFIABLE Mutex for shared state
	workerPool           *pond.WorkerPool                                 // Worker pool for downloading maps.
}

// Constructor for new downloader shared state
// NewDownloaderSharedState creates a new DownloaderSharedState.
func NewDownloaderSharedState(
	mapsDirectory string,
	existingMapFilesList []string,
	maxPoolCapacity int) (DownloaderSharedState, error) {

	mapHashAndTypeToName := make(map[string]string)
	for _, existingMap := range existingMapFilesList {
		englishMapName, err := readLocalizedDataFromMap(path.Join(mapsDirectory, existingMap))
		if err != nil {
			log.WithField("err", err).Error("Error reading map name from drive")
			return DownloaderSharedState{}, err
		}
		if englishMapName == "" {
			log.WithField("mapHashAndType", existingMap).
				Info("Map exists but map name is empty, should be removed and redownloaded.")
			return DownloaderSharedState{}, fmt.Errorf("map name was read but it is empty, should be removed and redownloaded.")
		}
		mapHashAndTypeToName[existingMap] = englishMapName
	}

	return DownloaderSharedState{
		mapDownloadDirectory: "maps",
		mapHashAndTypeToName: &mapHashAndTypeToName,
		currentlyDownloading: &map[string][]chan DownloadTaskReturnChannelInfo{},
		sharedRWMutex:        &sync.RWMutex{},
		workerPool:           pond.New(3, maxPoolCapacity, pond.Strategy(pond.Eager())),
	}, nil
}

// DownloadTaskState holds all of the information needed for the download task.
type DownloadTaskState struct {
	mapDownloadDirectory string
	mapHashAndType       string
	mapURL               url.URL
	mapHashAndTypeToName *map[string]string
	currentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo
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

	englishMapName, downloadTaskInfoChannel := dispatchMapDownloadTask(
		*downloaderSharedState,
		mapHashAndType,
		mapURL)

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

	// Reading the english map name from map file
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

	// Add to the mapHashAndTypeToName if the map is on drive:
	// Add the variable to the mapHashAndTypeToName
	// within the mutex lock to avoid IO operations while under lock.
	(*downloaderSharedState.mapHashAndTypeToName)[mapHashAndType] = englishMapName

	return englishMapName, nil
}

// downloadSingleMap handles downloading a single map based on an URL passed through
// the task state.
func downloadSingleMap(taskState DownloadTaskState) {
	log.WithField("taskState", taskState).Info("Entered downloadSingleMap()")

	outputFilepath := path.Join(taskState.mapDownloadDirectory, taskState.mapHashAndType)

	response, err := http.Get(taskState.mapURL.String())
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			"",
			fmt.Errorf("error downloading in http.Get map: %v", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			"",
			fmt.Errorf("error downloading, request returned code other than 200 OK"),
		)
		return
	}

	// Create output file:
	outFile, err := os.Create(outputFilepath)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			"",
			fmt.Errorf("error creating file in os.Create: %v", err))
		return
	}
	defer outFile.Close()

	// Copy contents of response to the file:
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			"",
			fmt.Errorf("error copying contents to file in io.Copy: %v", err))
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

// dispatchMapDownloadTask handles dispatching of the map download task, if
// the map is not available within the shared state under the mapHashAndTypeToName.
func dispatchMapDownloadTask(
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
		// if it is not then add key to the map and create one element
		// slice with the channel and submit the download task to the worker pool:
		(*downloaderSharedState.currentlyDownloading)[mapHashAndType] = []chan DownloadTaskReturnChannelInfo{downloadTaskInfoChannel}
		downloaderSharedState.workerPool.Submit(
			func() {
				// Errors are written to directly to the channel,
				// each of requesting goroutines will receive the error from
				// this function via the channel.
				downloadSingleMap(taskState)
			},
		)
	}
	return "", downloadTaskInfoChannel

}

// sendDownloadTaskReturnInfoToChannels iterates over all of the channels
// waiting for the download to finish, and sends the english map name or an error
// message through te channel.
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
