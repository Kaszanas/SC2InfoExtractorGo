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
// Set of dependencies that are currently being downloaded,
// this avoids starting multiple downloads of the same map:
// REVIEW: How to effectively wait for the download to finish
// REVIEW: if another thread requests the same map name?
// channel of string because the response is the map name:
// DownloaderSharedState holds all of the shared state for the downloader.
type DownloaderSharedState struct {
	// Directory where the maps will be downloaded:
	MapDownloadDirectory string // NOT_MODIFIABLE Directory where the maps are downloaded
	// Directory where other dependencies will be downloaded:
	DependencyDownloadDirectory string // NOT_MODIFIABLE Directory where other dependencies are downloaded
	// Set of dependencies that already exist on the drive:
	DownloadedDependenciesSet *map[string]struct{} // MODIFIABLE Mapping from filename to english map name
	// Map of dependencies that are currently being downloaded to the channels that are waiting for the download to finish:
	CurrentlyDownloading *map[string][]chan DownloadTaskReturnChannelInfo // MODIFIABLE Mapping from filename to list of channels to be notified when download finishes
	// Mutex for shared state:
	SharedRWMutex *sync.RWMutex // MODIFIABLE Mutex for shared state
	// Worker pool for downloading dependencies in parallel:
	WorkerPool *pond.WorkerPool // Worker pool for downloading dependencies.
}

// Constructor for new downloader shared state
// NewDownloaderSharedState creates a new DownloaderSharedState.
func NewDownloaderSharedState(
	cliFlags utils.CLIFlags,
) (DownloaderSharedState, error) {

	mapsDirectory := filepath.Join(
		cliFlags.DependencyDirectory,
		"maps",
	)

	existingFilesMapsSet, err := file_utils.ExistingFilesSet(
		mapsDirectory, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to get existing map files set.")
		return DownloaderSharedState{}, err
	}

	log.WithFields(log.Fields{
		"mapsDirectory":        mapsDirectory,
		"existingFilesMapsSet": len(existingFilesMapsSet)},
	).Debug("Entered NewDownloaderSharedState()")

	progressBar := utils.NewProgressBar(
		len(existingFilesMapsSet),
		"Initializing downloader: ",
	)

	for existingMapFilepath := range existingFilesMapsSet {

		_, err := sc2_map_processing.
			ReadLocalizedDataFromMapGetForeignToEnglishMapping(
				existingMapFilepath,
				progressBar,
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

	otherDependenciesDirectory := filepath.Join(
		cliFlags.DependencyDirectory,
		"other_dependencies",
	)

	otherDependencyFilesSet, err := file_utils.ExistingFilesSet(
		otherDependenciesDirectory, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).
			Error("Failed to get existing other dependencies files set.")
		return DownloaderSharedState{}, err
	}

	// Combine the two sets of existing files:
	dependenciesSet := make(map[string]struct{})
	for existingMapFilepath := range existingFilesMapsSet {
		dependenciesSet[existingMapFilepath] = struct{}{}
	}
	for existingOtherDependencyFilepath := range otherDependencyFilesSet {
		dependenciesSet[existingOtherDependencyFilepath] = struct{}{}
	}

	return DownloaderSharedState{
		MapDownloadDirectory:        mapsDirectory,
		DependencyDownloadDirectory: otherDependenciesDirectory,
		DownloadedDependenciesSet:   &dependenciesSet,
		CurrentlyDownloading:        &map[string][]chan DownloadTaskReturnChannelInfo{},
		SharedRWMutex:               &sync.RWMutex{},
		WorkerPool:                  pond.New(4, cliFlags.NumberOfThreads*2, pond.Strategy(pond.Eager())),
	}, nil
}

// DownloadTaskState holds all of the information needed for the download task.
type DownloadTaskState struct {
	dependencyDownloadDirectory string
	dependencyFilenameIsMap     sc2_map_processing.ReplayFilenameIsMap
	dependencyURL               url.URL
	downloadedDependenciesSet   *map[string]struct{}
	currentlyDownloading        *map[string][]chan DownloadTaskReturnChannelInfo
	sharedRWMutex               *sync.RWMutex
}

// DownloadTaskReturnChannelInfo holds the information needed to return
// after the download finishes.
type DownloadTaskReturnChannelInfo struct {
	err error
}

// TODO: Change name:
func DownloadDependencyIfNotExists(
	downloaderSharedState *DownloaderSharedState,
	filenameAndIsMap sc2_map_processing.ReplayFilenameIsMap,
	mapURL url.URL,
	progressBar *progressbar.ProgressBar,
) error {
	// Defer the progress bar increment:
	defer func() {
		if err := progressBar.Add(1); err != nil {
			log.WithField("error", err).
				Error("Error updating progress bar in DownloadDependencyIfNotExists")
		}
	}()

	log.WithFields(
		log.Fields{
			"dependencyFilename": filenameAndIsMap.DependencyFilename,
			"mapURL":             mapURL.String(),
		},
	).Debug("Entered DownloadDependencyIfNotExists()")

	downloadTaskInfoChannel := dispatchMapDownloadTask(
		*downloaderSharedState,
		filenameAndIsMap,
		mapURL,
	)
	if downloadTaskInfoChannel == nil {
		return nil
	}

	// Wait for channel to finish downloading the dependency.
	taskStatus := <-downloadTaskInfoChannel
	if taskStatus.err != nil {
		log.WithField("error", taskStatus.err).Error("Error downloading dependency")
		return fmt.Errorf("error downloading dependency: %v", taskStatus.err)
	}

	log.Debug("Finished DownloadDependencyIfNotExists()")
	return nil
}

// dispatchMapDownloadTask handles dispatching of the map download task, if
// the map is not available within the shared state under the mapHashAndExtensionToName.
func dispatchMapDownloadTask(
	downloaderSharedState DownloaderSharedState,
	filenameAndIsMap sc2_map_processing.ReplayFilenameIsMap,
	mapURL url.URL,
) chan DownloadTaskReturnChannelInfo {

	// Locking access to shared state:
	downloaderSharedState.SharedRWMutex.Lock()
	defer downloaderSharedState.SharedRWMutex.Unlock()

	// REVIEW: Is this the best way to go about it?
	// This is required because downloaded maps set contains full paths to the maps:

	maybeDependencyFilepath := ""
	switch filenameAndIsMap.IsMap {
	case true:
		maybeDependencyFilepath = filepath.Join(
			downloaderSharedState.MapDownloadDirectory,
			filenameAndIsMap.DependencyFilename,
		)
	case false:
		maybeDependencyFilepath = filepath.Join(
			downloaderSharedState.DependencyDownloadDirectory,
			filenameAndIsMap.DependencyFilename,
		)
	}

	// Check if the english map name was already read from the drive, return if present:
	_, ok := (*downloaderSharedState.DownloadedDependenciesSet)[maybeDependencyFilepath]
	if ok {
		log.WithField("DependencyFilename", filenameAndIsMap.DependencyFilename).
			Info("Dependency name was already processed in DownloadedDependenciesSet, returning.")
		return nil
	}

	// Create channel
	downloadTaskInfoChannel := make(chan DownloadTaskReturnChannelInfo)

	// Check if key is in currently downloading:
	listOfChannels, ok := (*downloaderSharedState.CurrentlyDownloading)[filenameAndIsMap.DependencyFilename]
	if ok {
		// If it is downloading then add the channel to the list of channels waiting for result
		// Map is being downloaded, add it to the list of currently downloading maps:
		log.WithField("DependencyFilename", filenameAndIsMap.DependencyFilename).
			Info("Dependency is being downloaded, adding channel to receive the result.")
		(*downloaderSharedState.CurrentlyDownloading)[filenameAndIsMap.DependencyFilename] =
			append(listOfChannels, downloadTaskInfoChannel)
	} else {
		log.WithField("DependencyFilename", filenameAndIsMap.DependencyFilename).
			Info("Dependency is not being downloaded, adding to download queue.")

		taskState := DownloadTaskState{
			downloadedDependenciesSet: downloaderSharedState.DownloadedDependenciesSet,
			currentlyDownloading:      downloaderSharedState.CurrentlyDownloading,
			dependencyFilenameIsMap:   filenameAndIsMap,
			dependencyURL:             mapURL,
			sharedRWMutex:             downloaderSharedState.SharedRWMutex,
		}

		switch filenameAndIsMap.IsMap {
		case true:
			taskState.dependencyDownloadDirectory = downloaderSharedState.MapDownloadDirectory
		case false:
			taskState.dependencyDownloadDirectory = downloaderSharedState.DependencyDownloadDirectory
		}

		// if it is not then add key to the map and create one element
		// slice with the channel and submit the download task to the worker pool:
		(*downloaderSharedState.CurrentlyDownloading)[filenameAndIsMap.DependencyFilename] =
			[]chan DownloadTaskReturnChannelInfo{downloadTaskInfoChannel}
		downloaderSharedState.WorkerPool.Submit(
			func() {
				// Errors are written to directly to the channel,
				// each of requesting goroutines will receive the error from
				// this function via the channel.
				downloadSingleDependency(taskState)
			},
		)
	}

	log.Debug("Finished dispatchMapDownloadTask()")
	return downloadTaskInfoChannel
}

// downloadSingleDependency handles downloading a single map based on an URL passed through
// the task state.
func downloadSingleDependency(taskState DownloadTaskState) {
	log.Debug("Entered downloadSingleDependency()")

	outputFilepath := filepath.Join(
		taskState.dependencyDownloadDirectory,
		taskState.dependencyFilenameIsMap.DependencyFilename,
	)

	response, err := http.Get(taskState.dependencyURL.String())
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			fmt.Errorf("error downloading in http.Get dependency: %v", err),
		)
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
			fmt.Errorf("error creating file in os.Create: %v", err),
		)
		return
	}
	defer outFile.Close()

	// Copy contents of response to the file:
	_, err = io.Copy(outFile, response.Body)
	if err != nil {
		sendDownloadTaskReturnInfoToChannels(
			&taskState,
			fmt.Errorf("error copying contents to file in io.Copy: %v", err),
		)
		return
	}

	sendDownloadTaskReturnInfoToChannels(&taskState, nil)
}

// sendDownloadTaskReturnInfoToChannels iterates over all of the channels
// waiting for the download to finish, and sends the english map name or an error
// message through te channel.
func sendDownloadTaskReturnInfoToChannels(
	taskState *DownloadTaskState,
	err error,
) {

	// Locking to ensure that other tasks do not read from the processed maps set.
	// this data structure acts as a source of truth to check what maps were already downloaded.
	// Initially it is only populated with the maps that are available on the drive.
	// But when downloading a map, this set is updated after a successful download.
	taskState.sharedRWMutex.Lock()
	defer taskState.sharedRWMutex.Unlock()

	(*taskState.downloadedDependenciesSet)[taskState.dependencyFilenameIsMap.DependencyFilename] = struct{}{}
	for _, channel := range (*taskState.currentlyDownloading)[taskState.dependencyFilenameIsMap.DependencyFilename] {
		channel <- DownloadTaskReturnChannelInfo{
			err: err,
		}
	}
	delete(*taskState.currentlyDownloading, taskState.dependencyFilenameIsMap.DependencyFilename)
}
