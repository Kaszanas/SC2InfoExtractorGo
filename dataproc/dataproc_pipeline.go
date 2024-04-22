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

	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

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
	cliFlags utils.CLIFlags,
) {

	log.Info("Entered PipelineWrapper()")
	// REVIEW: Start Review, New implementation of map translation below:
	mapsDirectoryName := "maps"
	// Create maps directory if it doesn't exist:
	mapsDirectory := utils.GetOrCreateMapsDirectory(mapsDirectoryName)
	if mapsDirectory == "" {
		return
	}
	existingMapFilesList := utils.ListFiles(mapsDirectory, ".s2ma")

	// Shared state for the downloader:
	downloaderSharedState := NewDownloaderSharedState(
		existingMapFilesList, cliFlags.NumberOfThreads*2)
	defer downloaderSharedState.workerPool.StopAndWait()
	// REVIEW: Finish Review

	// If it is specified by the user to perform the processing without multiprocessing GOMACPROCS needs to be set to 1 in order to allow 1 thread:
	runtime.GOMAXPROCS(cliFlags.NumberOfThreads)
	var channel = make(chan ReplayProcessingChannelContents, cliFlags.NumberOfThreads+1)
	var wg sync.WaitGroup
	// Adding a task for each of the supplied chunks to speed up the processing:
	wg.Add(cliFlags.NumberOfThreads)

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
					&downloaderSharedState,
					cliFlags,
				)
			}
		}()
	}

	for index, chunk := range fileChunks {
		channel <- ReplayProcessingChannelContents{Index: index, ChunkOfFiles: chunk}
	}

	close(channel)
	wg.Wait()

	log.Info("Finished PipelineWrapper()")
}

// MultiprocessingChunkPipeline is a single instance of processing that
// is meant to be spawned by the orchestrator in order to speed up the process of data extraction.
func MultiprocessingChunkPipeline(
	listOfFiles []string,
	packageToZipBool bool,
	compressionMethod uint16,
	chunkIndex int,
	downloaderSharedState *DownloaderSharedState,
	cliFlags utils.CLIFlags,
) {

	// Letting the orchestrator know that this processing task was finished:
	log.Info("Entered MultiprocessingChunkPipeline()")

	// Create ProcessingInfoFile:
	processingInfoFile, processingInfoStruct := utils.CreateProcessingInfoFile(
		cliFlags.LogFlags.LogPath,
		chunkIndex)
	defer processingInfoFile.Close()

	// Initializing grpc connection if the user chose to perform anonymization.
	grpcAnonymizer := checkAnonymizationInitializeGRPC(cliFlags.PerformChatAnonymization)
	// In order to free up resources We are defering the connection closing when all of the files have been processed:
	if grpcAnonymizer != nil {
		defer grpcAnonymizer.Connection.Close()
	}

	// Defining counters:
	pipelineErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0
	saveErrorCounter := 0

	// Helper method returning bytes buffer and zip writer which will be used to save the processing results into:
	var buffer *bytes.Buffer
	var writer *zip.Writer
	var packageSummary data.PackageSummary
	if packageToZipBool {
		buffer, writer = utils.InitBufferWriter()
		log.Info("Initialized buffer and writer.")

		// Create package summary structure:
		packageSummary = data.DefaultPackageSummary()
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
			downloaderSharedState,
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
			processingInfoStruct.FailedToProcess = append(
				processingInfoStruct.FailedToProcess,
				map[string]string{replayFile: failureReason})
			continue
		}

		// Saving output to zip archive:
		if packageToZipBool {
			// Append it to a list and when a package is created create a package summary and clear the list for next iterations
			data.AddReplaySummToPackageSumm(&replaySummary, &packageSummary)
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
			// TODO: This might be done easier.
			// Currently this is duplicate code and seems to introduce bad practice!
			processedCounter++
			processingInfoStruct.ProcessedFiles = append(
				processingInfoStruct.ProcessedFiles,
				replayFile)
			log.Info("Added file to zip archive.")
			continue
		}

		okSaveToDrive := utils.SaveFileToDrive(
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
		processingInfoStruct.ProcessedFiles = append(
			processingInfoStruct.ProcessedFiles,
			replayFile)
	}

	// Saving processingInfo to know which files failed to process:
	utils.SaveProcessingInfo(processingInfoFile, processingInfoStruct)
	log.Info("Saved processing.log")

	if packageToZipBool {
		// Writing PackageSummaryFile to drive:
		utils.CreatePackageSummaryFile(
			cliFlags.OutputDirectory,
			packageSummary,
			chunkIndex)

		// Writing the zip archive to drive:
		writer.Close()
		packageAbsPath := filepath.Join(
			cliFlags.OutputDirectory,
			"package_"+strconv.Itoa(chunkIndex)+".zip")
		err := os.WriteFile(packageAbsPath, buffer.Bytes(), 0777)
		if err != nil {
			log.WithFields(log.Fields{
				"packageAbsolutePath": packageAbsPath,
				"packageNumber":       chunkIndex}).
				Error("Failed to save package to drive!")
		}
	}

	log.Info("Finished MultiprocessingChunkPipeline()")

}

// FileProcessingPipeline is performing the whole data processing pipeline for a replay file. Reads the replay, cleans the replay structure, creates replay summary, anonymizes, and creates a JSON replay output.
func FileProcessingPipeline(
	replayFile string,
	grpcAnonymizer *GRPCAnonymizer,
	downloaderSharedState *DownloaderSharedState,
	cliFlags utils.CLIFlags,
) (bool, data.CleanedReplay, data.ReplaySummary, string) {

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
			data.CleanedReplay{},
			data.ReplaySummary{},
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
				data.CleanedReplay{},
				data.ReplaySummary{},
				fmt.Sprintf("checkIntegrity() failed: %s", failureReason)
		}
	}

	// Performing validity checks:
	if cliFlags.PerformValidityCheck {
		if cliFlags.FilterGameMode&Ranked1v1 != 0 && gameIs1v1Ranked(replayData) {
			// Perform Validity check
			if !validate1v1Replay(replayData) {
				return false,
					data.CleanedReplay{},
					data.ReplaySummary{},
					"validateReplay() failed"
			}
		}
	}

	// Filtering:
	if cliFlags.PerformFiltering {
		if !filterGameModes(replayData, cliFlags.FilterGameMode) {
			return false,
				data.CleanedReplay{},
				data.ReplaySummary{},
				"filterGameModes() failed"
		}
	}

	// Getting map URL and hash before mutexing, this operation is not thread safe:
	mapURL, mapHashAndType, ok := getMapURLAndHashFromReplayData(replayData)
	if !ok {
		return false, data.CleanedReplay{}, data.ReplaySummary{}, "getMapURLAndHashFromReplayData() failed"
	}
	// TODO: Check if the map is in the list of localized maps:

	// TODO: If it is not then download the map and add it to the list of localized maps:

	// REVIEW: Start Review, New implementation of map translation below:
	// Mutex start
	englishMapName := getEnglishMapNameDownloadIfNotExists(
		downloaderSharedState,
		mapHashAndType,
		mapURL)

	// Clean replay structure:
	cleanOk, cleanReplayStructure := extractReplayData(
		replayData,
		englishMapName,
		cliFlags.PerformCleanup)
	if !cleanOk {
		log.WithField("file", replayFile).Error("Failed to perform cleaning.")
		return false,
			data.CleanedReplay{},
			data.ReplaySummary{},
			"cleanReplay() failed"
	}
	// REVIEW: Finish Review

	// Create replay summary:
	summarizeOk, summarizedReplay := summarizeReplay(&cleanReplayStructure)
	if !summarizeOk {
		log.WithField("file", replayFile).Error("Failed to create replay summary.")
		return false,
			data.CleanedReplay{},
			data.ReplaySummary{},
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
				data.CleanedReplay{},
				data.ReplaySummary{},
				"anonymizeReplay() failed"
		}
	}

	log.Info("Finished FileProcessingPipeline()")

	return true, cleanReplayStructure, summarizedReplay, ""
}

// gameis1v1Ranked
func gameIs1v1Ranked(replayData *rep.Rep) bool {

	isAmm := replayData.InitData.GameDescription.GameOptions.Amm()
	isCompetitive := replayData.InitData.GameDescription.GameOptions.CompetitiveOrRanked()
	isTwoPlayers := len(replayData.Metadata.Players()) == 2
	return isAmm && isCompetitive && isTwoPlayers
}

// checkAnonymizationInitializeGRPC verifies if the anonymization should be performed and returns a pointer to GRPCAnonymizer.
func checkAnonymizationInitializeGRPC(
	performAnonymizationBool bool) *GRPCAnonymizer {
	if !performAnonymizationBool {
		return nil
	}

	log.Info("Detected that user wants anonymization, attempting to set up GRPCAnonymizer{}")
	grpcAnonymizer := GRPCAnonymizer{}
	if !grpcAnonymizer.grpcDialConnect() {
		log.Error("Could not connect to the gRPC server!")
	}
	grpcAnonymizer.grpcInitializeClient()
	grpcAnonymizer.Cache = make(map[string]string)

	return &grpcAnonymizer
}
