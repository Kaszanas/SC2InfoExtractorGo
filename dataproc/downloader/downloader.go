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
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	"github.com/alitto/pond"
	"github.com/schollz/progressbar/v3"
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
	DownloadedMapsSet    *map[string]struct{}                             // MODIFIABLE Mapping from filename to english map name
	CurrentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo // MODIFIABLE Mapping from filename to list of channels to be notified when download finishes
	SharedRWMutex        *sync.RWMutex                                    // MODIFIABLE Mutex for shared state
	WorkerPool           *pond.WorkerPool                                 // Worker pool for downloading maps.
}

// Constructor for new downloader shared state
// NewDownloaderSharedState creates a new DownloaderSharedState.
func NewDownloaderSharedState(
	cliFlags utils.CLIFlags,
) (DownloaderSharedState, error) {

	existingFilesMapsSet, err := file_utils.ExistingFilesSet(
		cliFlags.MapsDirectory, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to get existing map files set.")
		return DownloaderSharedState{}, err
	}

	log.WithFields(log.Fields{
		"mapsDirectory":        cliFlags.MapsDirectory,
		"existingFilesMapsSet": len(existingFilesMapsSet)}).
		Info("Entered NewDownloaderSharedState()")

	for existingMapFilepath := range existingFilesMapsSet {
		progressBarInitializeDownloader := utils.NewProgressBar(
			len(existingFilesMapsSet),
			"[0/4] Initializing downloader: ",
		)

		_, err := sc2_map_processing.
			ReadLocalizedDataFromMapGetForeignToEnglishMapping(
				existingMapFilepath,
				progressBarInitializeDownloader,
			)
		// if the map exists but cannot be read then it will automatically be
		// re-downloaded as it is not added to downloaded maps set
		if err != nil {
			log.WithField("error", err).
				Error("Error reading map name from drive. Map could not be processed")
			// File is removed to assure that
			// if it is corrupted it will be re-downloaded:
			delete(existingFilesMapsSet, existingMapFilepath)
			continue
		}
	}

	return DownloaderSharedState{
		MapDownloadDirectory: cliFlags.MapsDirectory,
		DownloadedMapsSet:    &existingFilesMapsSet,
		CurrentlyDownloading: &map[string][]chan DownloadTaskReturnChannelInfo{},
		SharedRWMutex:        &sync.RWMutex{},
		WorkerPool:           pond.New(4, cliFlags.NumberOfThreads*2, pond.Strategy(pond.Eager())),
	}, nil
}

// DownloadTaskState holds all of the information needed for the download task.
type DownloadTaskState struct {
	mapDownloadDirectory string
	mapHashAndExtension  string
	mapURL               url.URL
	processedMapsSet     *map[string]struct{}
	currentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo
	sharedRWMutex        *sync.RWMutex
}

// DownloadTaskReturnChannelInfo holds the information needed to return
// after the download finishes.
type DownloadTaskReturnChannelInfo struct {
	err error
}

// TODO: Change name:
func DownloadMapIfNotExists(
	downloaderSharedState *DownloaderSharedState,
	mapHashAndExtension string,
	mapURL url.URL,
	progressBar *progressbar.ProgressBar,
) error {
	// Defer the progress bar increment:
	defer func() {
		if err := progressBar.Add(1); err != nil {
			log.WithField("error", err).
				Error("Error updating progress bar in DownloadMapIfNotExists")
		}
	}()

	log.WithFields(
		log.Fields{
			"mapHashAndExtension": mapHashAndExtension,
			"mapURL":              mapURL.String(),
		},
	).Info("Entered getEnglishMapNameDownloadIfNotExists()")

	downloadTaskInfoChannel := dispatchMapDownloadTask(
		*downloaderSharedState,
		mapHashAndExtension,
		mapURL)
	if downloadTaskInfoChannel == nil {
		return nil
	}

	// Wait for channel to finish downloading the map.
	taskStatus := <-downloadTaskInfoChannel
	if taskStatus.err != nil {
		log.WithField("error", taskStatus.err).Error("Error downloading map")
		return fmt.Errorf("error downloading map: %v", taskStatus.err)
	}

	log.Info("Finished getEnglishMapNameDownloadIfNotExists()")
	return nil
}

// dispatchMapDownloadTask handles dispatching of the map download task, if
// the map is not available within the shared state under the mapHashAndExtensionToName.
func dispatchMapDownloadTask(
	downloaderSharedState DownloaderSharedState,
	mapHashAndExtension string,
	mapURL url.URL) chan DownloadTaskReturnChannelInfo {

	// Locking access to shared state:
	downloaderSharedState.SharedRWMutex.Lock()
	defer downloaderSharedState.SharedRWMutex.Unlock()

	// Check if the english map name was already read from the drive, return if present:
	_, ok := (*downloaderSharedState.DownloadedMapsSet)[mapHashAndExtension]
	if ok {
		log.WithField("mapHashAndExtension", mapHashAndExtension).
			Info("Map name was already processed in mapHashAndExtensionToName, returning.")
		return nil
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
		(*downloaderSharedState.CurrentlyDownloading)[mapHashAndExtension] =
			append(listOfChannels, downloadTaskInfoChannel)
	} else {
		log.WithField("mapHashAndExtension", mapHashAndExtension).
			Info("Map is not being downloaded, adding to download queue.")
		taskState := DownloadTaskState{
			mapDownloadDirectory: downloaderSharedState.MapDownloadDirectory,
			processedMapsSet:     downloaderSharedState.DownloadedMapsSet,
			currentlyDownloading: downloaderSharedState.CurrentlyDownloading,
			mapHashAndExtension:  mapHashAndExtension,
			mapURL:               mapURL,
			sharedRWMutex:        downloaderSharedState.SharedRWMutex,
		}
		// if it is not then add key to the map and create one element
		// slice with the channel and submit the download task to the worker pool:
		(*downloaderSharedState.CurrentlyDownloading)[mapHashAndExtension] =
			[]chan DownloadTaskReturnChannelInfo{downloadTaskInfoChannel}
		downloaderSharedState.WorkerPool.Submit(
			func() {
				// Errors are written to directly to the channel,
				// each of requesting goroutines will receive the error from
				// this function via the channel.
				downloadSingleMap(taskState)
			},
		)
	}

	log.Info("Finished dispatchMapDownloadTask()")
	return downloadTaskInfoChannel
}

// downloadSingleMap handles downloading a single map based on an URL passed through
// the task state.
func downloadSingleMap(taskState DownloadTaskState) {
	log.Info("Entered downloadSingleMap()")

	outputFilepath := filepath.Join(
		taskState.mapDownloadDirectory,
		taskState.mapHashAndExtension,
	)

	response, err := http.Get(taskState.mapURL.String())
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			fmt.Errorf("error downloading in http.Get map: %v", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			fmt.Errorf("error downloading, request returned code other than 200 OK"),
		)
		return
	}

	// Create output file:
	outFile, err := os.Create(outputFilepath)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			fmt.Errorf("error creating file in os.Create: %v", err))
		return
	}
	defer outFile.Close()

	// Copy contents of response to the file:
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			fmt.Errorf("error copying contents to file in io.Copy: %v", err))
		return
	}

	sendDownloadTaskReturnInfoToChannels(&taskState, nil)
}

// sendDownloadTaskReturnInfoToChannels iterates over all of the channels
// waiting for the download to finish, and sends the english map name or an error
// message through te channel.
func sendDownloadTaskReturnInfoToChannels(
	taskState *DownloadTaskState,
	err error) {
	taskState.sharedRWMutex.Lock()
	defer taskState.sharedRWMutex.Unlock()

	// TODO: This needs to change, I am no longer using the map hash and extension as the key
	// The mapHashAndExtensionToName has to change into a set implementation:
	(*taskState.processedMapsSet)[taskState.mapHashAndExtension] = struct{}{}
	for _, channel := range (*taskState.currentlyDownloading)[taskState.mapHashAndExtension] {
		channel <- DownloadTaskReturnChannelInfo{
			err: err,
		}
	}
	delete(*taskState.currentlyDownloading, taskState.mapHashAndExtension)
}
