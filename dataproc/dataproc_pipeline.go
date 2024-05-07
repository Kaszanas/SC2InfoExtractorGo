package dataproc

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/sc2_map_processing"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// ReplayProcessingChannelContents is a struct that is used to pass data
// between the orchestrator and the workers in the pipeline.
type ReplayProcessingChannelContents struct {
	Index        int
	ChunkOfFiles []string
}

// PipelineWrapper is an orchestrator that distributes work
// among available workers (threads)
func PipelineWrapper(
	fileChunks [][]string,
	packageToZipBool bool,
	compressionMethod uint16,
	mapsDirectoryPath string,
	processedReplaysFilepath string,
	cliFlags utils.CLIFlags,
) {

	log.Info("Entered PipelineWrapper()")
	// Create maps directory if it doesn't exist:
	err := file_utils.GetOrCreateDirectory(mapsDirectoryPath)
	if err != nil {
		log.WithField("error", err).Error("Failed to create maps directory.")
		return
	}

	// REVIEW: Start Review:
	existingMapFilesSet, err := file_utils.ExistingFilesSet(
		mapsDirectoryPath, ".s2ma",
	)
	if err != nil {
		log.WithField("error", err).Error("Failed to get existing map files set.")
		return
	}

	// Shared state for the downloader:
	downloadedMapFilesSet := make(map[string]struct{})
	downloaderSharedState, err := downloader.NewDownloaderSharedState(
		mapsDirectoryPath,
		existingMapFilesSet,
		downloadedMapFilesSet,
		cliFlags.NumberOfThreads*2)
	defer downloaderSharedState.WorkerPool.StopAndWait()
	if err != nil {
		log.WithField("error", err).Error("Failed to create downloader shared state.")
		return
	}

	// STAGE ONE PRE-PROCESS:
	// Download all SC2 maps from the replays if they were not processed before:
	existingMapFilesSet, err = donwloadAllSC2MapsHandleProcessedReplays(
		&downloaderSharedState,
		mapsDirectoryPath,
		processedReplaysFilepath,
		fileChunks,
		cliFlags,
	)
	if err != nil {
		log.WithField("error", err).Error("Failed to download all SC2 maps.")
		return
	}

	// STAGE TWO PRE-PROCESS:
	// Read all of the map names from the drive and create a mapping
	// from foreign to english names:
	mainForeignToEnglishMapping := make(map[string]string)
	for existingMapFilepath := range existingMapFilesSet {
		foreignToEnglishMapping, err := sc2_map_processing.
			ReadLocalizedDataFromMapGetForeignToEnglishMapping(
				existingMapFilepath,
			)
		if err != nil {
			log.WithField("error", err).
				Error("Error reading map name from drive. Map could not be processed")
			return
		}

		// Fill out the mapping, these maps won't be opened again:
		for foreignName, englishName := range foreignToEnglishMapping {
			mainForeignToEnglishMapping[foreignName] = englishName
		}
	}

	// REVIEW: Finish Review

	// If it is specified by the user to perform the processing without
	// multiprocessing GOMACPROCS needs to be set to 1 in order to allow 1 thread:
	runtime.GOMAXPROCS(cliFlags.NumberOfThreads)
	var channel = make(chan ReplayProcessingChannelContents, cliFlags.NumberOfThreads+1)
	var wg sync.WaitGroup
	// Adding a task for each of the supplied chunks to speed up the processing:
	wg.Add(cliFlags.NumberOfThreads)

	// Spin up workers waiting for chunks to process:
	for i := 0; i < cliFlags.NumberOfThreads; i++ {
		go func() {
			for {
				channelContents, ok := <-channel
				if !ok {
					wg.Done()
					return
				}
				MultiprocessingChunkPipeline(
					channelContents.ChunkOfFiles,
					packageToZipBool,
					compressionMethod,
					channelContents.Index,
					mainForeignToEnglishMapping,
					cliFlags,
				)
			}
		}()
	}

	// Passing the chunks to the workers:
	for index, chunk := range fileChunks {
		channel <- ReplayProcessingChannelContents{Index: index, ChunkOfFiles: chunk}
	}

	close(channel)
	wg.Wait()

	log.Info("Finished PipelineWrapper()")
}

// MultiprocessingChunkPipeline is a single instance of processing that
// is meant to be spawned by the orchestrator
// in order to speed up the process of data extraction.
func MultiprocessingChunkPipeline(
	listOfFiles []string,
	packageToZipBool bool,
	compressionMethod uint16,
	chunkIndex int,
	englishToForeignMapping map[string]string,
	cliFlags utils.CLIFlags,
) {

	// Letting the orchestrator know that this processing task was finished:
	log.Info("Entered MultiprocessingChunkPipeline()")

	// Create ProcessingInfoFile:
	processingInfoFile, processingInfoStruct, err := persistent_data.CreateProcessingInfoFile(
		cliFlags.LogFlags.LogPath,
		chunkIndex)
	if err != nil {
		log.WithField("error", err).Error("Failed to create processingInfoFile.")
		return
	}
	defer processingInfoFile.Close()

	// Initializing grpc connection if the user chose to perform anonymization.
	grpcAnonymizer := checkAnonymizationInitializeGRPC(cliFlags.PerformChatAnonymization)
	// In order to free up resources We are defering the connection closing when
	// all of the files have been processed:
	if grpcAnonymizer != nil {
		defer grpcAnonymizer.Connection.Close()
	}

	// TODO: These could be a separate data structure:
	// Defining counters:
	pipelineErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0
	saveErrorCounter := 0

	// Helper method returning bytes buffer and zip writer which will be
	// used to save the processing results into:
	var buffer *bytes.Buffer
	var writer *zip.Writer
	var packageSummary persistent_data.PackageSummary
	if packageToZipBool {
		buffer, writer = utils.InitBufferWriter()
		log.Info("Initialized buffer and writer.")

		// Create package summary structure:
		packageSummary = persistent_data.NewPackageSummary()
	}

	// Processing file:
	for _, replayFile := range listOfFiles {
		// Checking if the file was previously processed:
		if contains(processingInfoStruct.ProcessedFiles, replayFile) {
			continue
		}

		// Running all of the processing logic and verifying if it worked:
		didWork, cleanReplayStructure, replaySummary, failureReason := FileProcessingPipeline(
			replayFile,
			grpcAnonymizer,
			englishToForeignMapping,
			cliFlags,
		)

		// Create final replay string:
		stringifyOk, replayString := stringifyReplay(&cleanReplayStructure)
		if !stringifyOk {
			log.WithField("file", replayFile).
				Error("Failed to stringify the replay.")
			continue
		}

		if !didWork {
			pipelineErrorCounter++
			log.WithFields(log.Fields{
				"pipelineErrorCounter": pipelineErrorCounter,
				"replayFile":           replayFile,
			}).Error("Failed to perform FileProcessingPipeline()!")
			processingInfoStruct.AddToFailed(
				replayFile,
				failureReason,
			)
			continue
		}

		// Saving output to zip archive:
		if packageToZipBool {
			// Append it to a list and when a package is created create a package summary and clear the list for next iterations
			persistent_data.AddReplaySummToPackageSumm(
				&replaySummary,
				&packageSummary,
			)
			log.Info("Added replaySummary to packageSummary")

			savedSuccess := utils.SaveFileToArchive(
				replayString,
				replayFile,
				compressionMethod,
				writer)
			if !savedSuccess {
				compressionErrorCounter++
				log.WithFields(log.Fields{
					"compressionErrorCounter": compressionErrorCounter,
					"replayFile":              replayFile,
				}).Error("Failed to save file to archive! Skipping.")
				continue
			}

			processedCounter++
			processingInfoStruct.AddToProcessed(replayFile)
			log.Info("Added file to zip archive.")
			continue
		}

		okSaveToDrive := file_utils.SaveReplayJSONFileToDrive(
			replayString,
			replayFile,
			cliFlags.OutputDirectory)
		if !okSaveToDrive {
			saveErrorCounter++
			log.WithFields(log.Fields{
				"replayFile":               replayFile,
				"cliFlags.OutputDirectory": cliFlags.OutputDirectory,
				"saveErrorCounter":         saveErrorCounter,
			}).Error("Failed to save .json to drive!")
			continue
		}

		processedCounter++
		replayFileNameAndExtension := filepath.Base(replayFile)
		processingInfoStruct.AddToProcessed(replayFileNameAndExtension)
	}

	// Saving processingInfo to know which files failed to process:
	persistent_data.SaveProcessingInfoToFile(
		processingInfoFile,
		processingInfoStruct,
	)
	log.Info("Saved processing.log")

	if packageToZipBool {
		// Writing PackageSummaryFile to drive:
		persistent_data.CreatePackageSummaryFile(
			cliFlags.OutputDirectory,
			packageSummary,
			chunkIndex)

		// Writing the zip archive to drive:
		writer.Close()
		packagePath := filepath.Join(
			cliFlags.OutputDirectory,
			"package_"+strconv.Itoa(chunkIndex)+".zip")
		packageAbsPath, err := filepath.Abs(packagePath)
		if err != nil {
			log.WithFields(log.Fields{
				"packagePath":   packagePath,
				"packageNumber": chunkIndex}).
				Error("Failed to get absolute path of package!")
		}
		err = os.WriteFile(packageAbsPath, buffer.Bytes(), 0777)
		if err != nil {
			log.WithFields(log.Fields{
				"packageAbsolutePath": packageAbsPath,
				"packageNumber":       chunkIndex}).
				Error("Failed to save package to drive!")
		}
	}

	log.Info("Finished MultiprocessingChunkPipeline()")
}

// FileProcessingPipeline is performing the whole data processing pipeline
// for a replay file. Reads the replay, cleans the replay structure,
// creates replay summary, anonymizes, and creates a JSON replay output.
func FileProcessingPipeline(
	replayFile string,
	grpcAnonymizer *GRPCAnonymizer,
	englishToForeignMapping map[string]string,
	cliFlags utils.CLIFlags,
) (bool, replay_data.CleanedReplay, persistent_data.ReplaySummary, string) {

	log.Info("Entered FileProcessingPipeline()")

	// Read replay:
	replayData, err := rep.NewFromFile(replayFile)
	if err != nil {
		log.WithFields(log.Fields{
			"file":      replayFile,
			"error":     err,
			"readError": true}).
			Error("Failed to read file.")
		return false,
			replay_data.CleanedReplay{},
			persistent_data.ReplaySummary{},
			"rep.NewFromFile() failed"
	}
	log.WithField("file", replayFile).Info("Read data from a replay.")
	defer replayData.Close()

	// Performing integrity checks:
	if cliFlags.PerformIntegrityCheck {
		integrityOk, failureReason := checkIntegrity(replayData)
		if !integrityOk {
			log.WithField("file", replayFile).
				Error("Integrity check failed in file.")
			return false,
				replay_data.CleanedReplay{},
				persistent_data.ReplaySummary{},
				fmt.Sprintf("checkIntegrity() failed: %s", failureReason)
		}
	}

	// Performing validity checks:
	if cliFlags.PerformValidityCheck {
		if cliFlags.FilterGameMode&Ranked1v1 != 0 && gameIs1v1Ranked(replayData) {
			// Perform Validity check
			if !validate1v1Replay(replayData) {
				return false,
					replay_data.CleanedReplay{},
					persistent_data.ReplaySummary{},
					"validateReplay() failed"
			}
		}
	}

	// Filtering:
	if cliFlags.PerformFiltering {
		if !filterGameModes(replayData, cliFlags.FilterGameMode) {
			return false,
				replay_data.CleanedReplay{},
				persistent_data.ReplaySummary{},
				"filterGameModes() failed"
		}
	}

	// REVIEW: Start Review, New implementation of map translation below:
	// Clean replay structure:
	cleanOk, cleanReplayStructure := extractReplayData(
		replayData,
		englishToForeignMapping,
		cliFlags.PerformCleanup)
	if !cleanOk {
		log.WithField("file", replayFile).Error("Failed to perform cleaning.")
		return false,
			replay_data.CleanedReplay{},
			persistent_data.ReplaySummary{},
			"cleanReplay() failed"
	}
	// REVIEW: Finish Review

	// Create replay summary:
	summarizeOk, summarizedReplay := summarizeReplay(&cleanReplayStructure)
	if !summarizeOk {
		log.WithField("file", replayFile).Error("Failed to create replay summary.")
		return false,
			replay_data.CleanedReplay{},
			persistent_data.ReplaySummary{},
			"summarizeReplay() failed"
	}

	// Anonymize replay:
	if grpcAnonymizer != nil {
		if !anonymizeReplay(
			&cleanReplayStructure,
			grpcAnonymizer,
			cliFlags.PerformChatAnonymization) {
			log.WithField("file", replayFile).
				Error("Failed to anonymize replay.")
			return false,
				replay_data.CleanedReplay{},
				persistent_data.ReplaySummary{},
				"anonymizeReplay() failed"
		}
	}

	log.Info("Finished FileProcessingPipeline()")

	return true, cleanReplayStructure, summarizedReplay, ""
}

// gameis1v1Ranked checks if the replay is a 1v1 ranked game.
func gameIs1v1Ranked(replayData *rep.Rep) bool {

	isAmm := replayData.InitData.GameDescription.GameOptions.Amm()
	isCompetitive := replayData.InitData.GameDescription.GameOptions.CompetitiveOrRanked()
	isTwoPlayers := len(replayData.Metadata.Players()) == 2
	return isAmm && isCompetitive && isTwoPlayers
}
