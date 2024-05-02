package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/alitto/pond"
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
	MapDownloadDirectory string // NOT_MODIFIABLE Directory where the maps are downloaded
	// TODO: This needs to change, I am no longer using the map hash and extension as the key
	// The mapHashAndExtensionToName has to change into a set implementation:
	MapHashAndExtensionToName *map[string]string                               // MODIFIABLE Mapping from filename to english map name
	CurrentlyDownloading      *map[string][]chan DownloadTaskReturnChannelInfo // MODIFIABLE Mapping from filename to list of channels to be notified when download finishes
	SharedRWMutex             *sync.RWMutex                                    // MODIFIABLE Mutex for shared state
	WorkerPool                *pond.WorkerPool                                 // Worker pool for downloading maps.
}

// Constructor for new downloader shared state
// NewDownloaderSharedState creates a new DownloaderSharedState.
func NewDownloaderSharedState(
	mapsDirectory string,
	mapHashAndExtensionToName map[string]string,
	existingMapFilesList []string,
	maxPoolCapacity int) (DownloaderSharedState, error) {

	log.WithFields(log.Fields{
		"mapsDirectory":          mapsDirectory,
		"lenExistingMapFileList": len(existingMapFilesList)}).
		Info("Entered NewDownloaderSharedState()")

	for _, existingMapFilepath := range existingMapFilesList {
		foreighToEnglishMapName, err := sc2_map_processing.ReadLocalizedDataFromMapGetForeignToEnglishMapping(
			existingMapFilepath,
		)
		if err != nil {
			log.WithField("error", err).
				Error("Error reading map name from drive.")
			return DownloaderSharedState{},
				fmt.Errorf("map name was read but map is empty, should be removed and redownloaded: %s", err)
		}

		// TODO: This has to be merging multiple maps together:
		// I am no longer using the map hash and extension as the key
		// This is because chinese servers do not exist
		// any more and multiple different hashes
		// can point to the same map.
		mapFileBase := filepath.Base(existingMapFilepath)
		mapHashAndExtensionToName[mapFileBase] = foreighToEnglishMapName
	}

	return DownloaderSharedState{
		MapDownloadDirectory:      mapsDirectory,
		MapHashAndExtensionToName: &mapHashAndExtensionToName,
		CurrentlyDownloading:      &map[string][]chan DownloadTaskReturnChannelInfo{},
		SharedRWMutex:             &sync.RWMutex{},
		WorkerPool:                pond.New(4, maxPoolCapacity, pond.Strategy(pond.Eager())),
	}, nil
}

// DownloadTaskState holds all of the information needed for the download task.
type DownloadTaskState struct {
	mapDownloadDirectory      string
	mapHashAndExtension       string
	mapURL                    url.URL
	mapHashAndExtensionToName *map[string]string
	currentlyDownloading      *map[string][]chan DownloadTaskReturnChannelInfo
	sharedRWMutex             *sync.RWMutex
}

// DownloadTaskReturnChannelInfo holds the information needed to return after the download finishes.
type DownloadTaskReturnChannelInfo struct {
	mapNameString string
	err           error
}

func GetEnglishMapNameDownloadIfNotExists(
	downloaderSharedState *DownloaderSharedState,
	mapHashAndExtension string,
	mapURL url.URL) string {
	log.WithFields(
		log.Fields{
			"downloaderSharedState": downloaderSharedState,
			"mapHashAndExtension":   mapHashAndExtension,
			"mapURL":                mapURL.String(),
		},
	).Info("Entered getEnglishMapNameDownloadIfNotExists()")

	englishMapName, downloadTaskInfoChannel := dispatchMapDownloadTask(
		*downloaderSharedState,
		mapHashAndExtension,
		mapURL)

	if englishMapName != "" {
		return englishMapName
	}

	// Wait for channel to finish downloading the map.
	taskStatus := <-downloadTaskInfoChannel
	if taskStatus.err != nil {
		log.WithField("error", taskStatus.err).Error("Error downloading map")
		return ""
	}

	log.Info("Finished getEnglishMapNameDownloadIfNotExists()")
	return taskStatus.mapNameString
}

// dispatchMapDownloadTask handles dispatching of the map download task, if
// the map is not available within the shared state under the mapHashAndExtensionToName.
func dispatchMapDownloadTask(
	downloaderSharedState DownloaderSharedState,
	mapHashAndExtension string,
	mapURL url.URL) (string, chan DownloadTaskReturnChannelInfo) {

	// Locking access to shared state:
	downloaderSharedState.SharedRWMutex.Lock()
	defer downloaderSharedState.SharedRWMutex.Unlock()

	// Check if the english map name was already read from the drive, return if present:
	englishMapName, ok := (*downloaderSharedState.MapHashAndExtensionToName)[mapHashAndExtension]
	if ok {
		log.WithField("englishMapName", englishMapName).
			Info("Map name already exists in mapHashAndExtensionToName, returning.")
		return englishMapName, nil
	}

	// Create channel
	downloadTaskInfoChannel := make(chan DownloadTaskReturnChannelInfo)

	// Check if key is in currently downloading:
	listOfChannels, ok := (*downloaderSharedState.CurrentlyDownloading)[mapHashAndExtension]
	if ok {
		// If it is downloading then add the channel to the list of channels waiting for result
		// Map is being downloaded, add it to the list of currently downloading maps:
		log.WithField("mapHashAndExtension", mapHashAndExtension).
			Info("Map is being downloaded, adding channel to receive the result.")
		(*downloaderSharedState.CurrentlyDownloading)[mapHashAndExtension] = append(listOfChannels, downloadTaskInfoChannel)
	} else {
		log.WithField("mapHashAndExtension", mapHashAndExtension).
			Info("Map is not being downloaded, adding to download queue.")
		taskState := DownloadTaskState{
			mapDownloadDirectory:      downloaderSharedState.MapDownloadDirectory,
			mapHashAndExtensionToName: downloaderSharedState.MapHashAndExtensionToName,
			currentlyDownloading:      downloaderSharedState.CurrentlyDownloading,
			mapHashAndExtension:       mapHashAndExtension,
			mapURL:                    mapURL,
			sharedRWMutex:             downloaderSharedState.SharedRWMutex,
		}
		// if it is not then add key to the map and create one element
		// slice with the channel and submit the download task to the worker pool:
		(*downloaderSharedState.CurrentlyDownloading)[mapHashAndExtension] = []chan DownloadTaskReturnChannelInfo{downloadTaskInfoChannel}
		downloaderSharedState.WorkerPool.Submit(
			func() {
				// Errors are written to directly to the channel,
				// each of requesting goroutines will receive the error from
				// this function via the channel.
				downloadSingleMap(taskState)
			},
		)
	}

	log.WithField("downloadTaskInfoChannel", downloadTaskInfoChannel).
		Info("Finished dispatchMapDownloadTask()")
	return "", downloadTaskInfoChannel
}

// downloadSingleMap handles downloading a single map based on an URL passed through
// the task state.
func downloadSingleMap(taskState DownloadTaskState) {
	log.WithField("taskState", taskState).Info("Entered downloadSingleMap()")

	outputFilepath := filepath.Join(
		taskState.mapDownloadDirectory,
		taskState.mapHashAndExtension,
	)

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

	// TODO: Create another structure and
	englishMapName, err := sc2_map_processing.ReadLocalizedDataFromMapGetForeignToEnglishMapping(outputFilepath)
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

// sendDownloadTaskReturnInfoToChannels iterates over all of the channels
// waiting for the download to finish, and sends the english map name or an error
// message through te channel.
func sendDownloadTaskReturnInfoToChannels(
	taskState *DownloadTaskState,
	englishMapName string,
	err error) {
	taskState.sharedRWMutex.Lock()
	defer taskState.sharedRWMutex.Unlock()

	// TODO: This needs to change, I am no longer using the map hash and extension as the key
	// The mapHashAndExtensionToName has to change into a set implementation:
	(*taskState.mapHashAndExtensionToName)[taskState.mapHashAndExtension] = englishMapName
	for _, channel := range (*taskState.currentlyDownloading)[taskState.mapHashAndExtension] {
		channel <- DownloadTaskReturnChannelInfo{
			mapNameString: englishMapName,
			err:           err,
		}
	}
	delete(*taskState.currentlyDownloading, taskState.mapHashAndExtension)
}
