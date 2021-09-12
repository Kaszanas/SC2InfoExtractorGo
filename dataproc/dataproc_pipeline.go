package dataproc

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

type ChannelContents struct {
	Index int
	Chunk []string
}

// PipelineWrapper is an orchestrator that distributes work among available workers (threads)
func PipelineWrapper(absolutePathOutputDirectory string,
	chunks [][]string,
	packageToZipBool bool,
	performIntegrityCheckBool bool,
	performValidityCheckBool bool,
	gameModeCheckFlag int,
	performPlayerAnonymizationBool bool,
	performChatAnonymizationBool bool,
	performCleanupBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	numberOfThreads int,
	logsFilepath string) {

	log.Info("Entered PipelineWrapper()")

	// If it is specified by the user to perform the processing without multiprocessing GOMACPROCS needs to be set to 1 in order to allow 1 thread:
	runtime.GOMAXPROCS(numberOfThreads)

	var channel = make(chan ChannelContents, numberOfThreads+1)
	var wg sync.WaitGroup

	// Adding a task for each of the supplied chunks to speed up the processing:
	wg.Add(numberOfThreads)

	for i := 0; i < numberOfThreads; i++ {
		go func() {
			for {
				channelContents, ok := <-channel
				if !ok {
					wg.Done()
					return
				}
				MultiprocessingChunkPipeline(
					absolutePathOutputDirectory,
					channelContents.Chunk,
					packageToZipBool,
					performIntegrityCheckBool,
					performValidityCheckBool,
					gameModeCheckFlag,
					performPlayerAnonymizationBool,
					performChatAnonymizationBool,
					performCleanupBool,
					localizedMapsMap,
					compressionMethod,
					logsFilepath,
					channelContents.Index)
			}
		}()
	}

	for index, chunk := range chunks {
		channel <- ChannelContents{Index: index, Chunk: chunk}
	}

	close(channel)
	wg.Wait()

	log.Info("Finished PipelineWrapper()")
}

// MultiprocessingChunkPipeline is a single instance of processing that is meant to be spawned by the orchestrator in order to speed up the process of data extraction.
func MultiprocessingChunkPipeline(
	absolutePathOutputDirectory string,
	listOfFiles []string,
	packageToZipBool bool,
	performIntegrityCheckBool bool,
	performValidityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performChatAnonymizationBool bool,
	performCleanupBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	logsFilepath string,
	chunkIndex int) {

	// Letting the orchestrator know that this processing task was finished:
	log.Info("Entered MultiprocessingChunkPipeline()")

	// Create ProcessingInfoFile:
	processingInfoFile, processingInfoStruct := utils.CreateProcessingInfoFile(logsFilepath, chunkIndex)
	defer processingInfoFile.Close()

	// Initializing grpc connection if the user chose to perform anonymization.
	grpcAnonymizer := checkAnonymizationInitializeGRPC(performAnonymizationBool)
	// In order to free up resources We are defering the connection closing when all of the files have been processed:
	if grpcAnonymizer != nil {
		defer grpcAnonymizer.Connection.Close()
	}

	// Defining counters:
	pipelineErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0

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
		didWork, replayString, replaySummary, failureReason := FileProcessingPipeline(
			replayFile,
			performIntegrityCheckBool,
			performValidityCheckBool,
			gameModeCheckFlag,
			grpcAnonymizer,
			performChatAnonymizationBool,
			performCleanupBool,
			localizedMapsMap)

		if !didWork {
			pipelineErrorCounter++
			log.WithFields(log.Fields{
				"pipelineErrorCounter": pipelineErrorCounter,
				"replayFile":           replayFile,
			}).Error("Failed to perform FileProcessingPipeline()!")
			processingInfoStruct.FailedToProcess = append(processingInfoStruct.FailedToProcess, map[string]string{replayFile: failureReason})
			continue
		}

		// Saving output to zip archive:
		if packageToZipBool {
			// Append it to a list and when a package is created create a package summary and clear the list for next iterations
			data.AddReplaySummToPackageSumm(&replaySummary, &packageSummary)
			log.Info("Added replaySummary to packageSummary")

			savedSuccess := utils.SaveFileToArchive(replayString, replayFile, compressionMethod, writer)
			if !savedSuccess {
				compressionErrorCounter++
				log.WithFields(log.Fields{
					"compressionErrorCounter": compressionErrorCounter,
					"replayFile":              replayFile,
				}).Error("Failed to save file to archive! Skipping.")
				continue
			}
			log.Info("Added file to zip archive.")
		}

		processedCounter++
		processingInfoStruct.ProcessedFiles = append(processingInfoStruct.ProcessedFiles, replayFile)
	}

	// Saving processingInfo to know which files failed to process:
	utils.SaveProcessingInfo(processingInfoFile, processingInfoStruct)
	log.Info("Saved processing.log")

	// Writing PackageSummaryFile to drive:
	utils.CreatePackageSummaryFile(absolutePathOutputDirectory, packageSummary, chunkIndex)

	// Writing the zip archive to drive:
	writer.Close()
	packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(chunkIndex)+".zip")
	err := ioutil.WriteFile(packageAbsPath, buffer.Bytes(), 0777)
	if err != nil {
		log.WithFields(log.Fields{
			"packageAbsolutePath": packageAbsPath,
			"packageNumber":       chunkIndex}).Error("Failed to save package to drive!")
	}

	log.Info("Finished MultiprocessingChunkPipeline()")

}

// FileProcessingPipeline is performing the whole data processing pipeline for a replay file. Reads the replay, cleans the replay structure, creates replay summary, anonymizes, and creates a JSON replay output.
func FileProcessingPipeline(replayFile string,
	performIntegrityCheckBool bool,
	performValidityCheckBool bool,
	gameModeCheckFlag int,
	grpcAnonymizer *GRPCAnonymizer,
	performChatAnonymizationBool bool,
	performCleanupBool bool,
	localizedMapsMap map[string]interface{}) (bool, string, data.ReplaySummary, string) {

	log.Info("Entered FileProcessingPipeline()")

	// Read replay:
	replayData, err := rep.NewFromFile(replayFile)
	if err != nil {
		log.WithFields(log.Fields{"file": replayFile, "error": err, "readError": true}).Error("Failed to read file.")
		return false, "", data.ReplaySummary{}, "rep.NewFromFile() failed"
	}
	log.WithField("file", replayFile).Info("Read data from a replay.")

	// Performing integrity checks
	if performIntegrityCheckBool && !checkIntegrity(replayData) {
		log.WithField("file", replayFile).Error("Integrity check failed in file.")
		return false, "", data.ReplaySummary{}, "checkIntegrity() failed"
	}

	// Performing validity checks:
	if performValidityCheckBool {
		if gameModeCheckFlag&Ranked1v1 != 0 && gameIs1v1Ranked(replayData) {
			// Perform Validity check
			if !validate1v1Replay(replayData) {
				return false, "", data.ReplaySummary{}, "validateReplay() failed"
			}
		}
	}

	// Filtering
	if !filterGameModes(replayData, gameModeCheckFlag) {
		return false, "", data.ReplaySummary{}, "filterGameModes() failed"
	}

	// Clean replay structure:
	cleanOk, cleanReplayStructure := extractReplayData(replayData, localizedMapsMap, performCleanupBool)
	if !cleanOk {
		log.WithField("file", replayFile).Error("Failed to perform cleaning.")
		return false, "", data.ReplaySummary{}, "cleanReplay() failed"
	}

	// Create replay summary:
	summarizeOk, summarizedReplay := summarizeReplay(&cleanReplayStructure)
	if !summarizeOk {
		log.WithField("file", replayFile).Error("Failed to create replay summary.")
		return false, "", data.ReplaySummary{}, "summarizeReplay() failed"
	}

	// Anonimize replay:
	if grpcAnonymizer != nil {
		if !anonymizeReplay(&cleanReplayStructure, grpcAnonymizer, performChatAnonymizationBool) {
			log.WithField("file", replayFile).Error("Failed to anonymize replay.")
			return false, "", data.ReplaySummary{}, "anonymizeReplay() failed"
		}
	}

	// Create final replay string:
	stringifyOk, finalReplayString := stringifyReplay(&cleanReplayStructure)
	if !stringifyOk {
		log.WithField("file", replayFile).Error("Failed to stringify the replay.")
		return false, "", data.ReplaySummary{}, "stringifyReplay() failed"
	}

	replayData.Close()

	log.Info("Finished FileProcessingPipeline()")

	return true, finalReplayString, summarizedReplay, ""
}

// gameis1v1Ranked
func gameIs1v1Ranked(replayData *rep.Rep) bool {

	isAmm := replayData.InitData.GameDescription.GameOptions.Amm()
	isCompetitive := replayData.InitData.GameDescription.GameOptions.CompetitiveOrRanked()
	isTwoPlayers := len(replayData.Metadata.Players()) == 2
	return isAmm && isCompetitive && isTwoPlayers
}

// checkAnonymizationInitializeGRPC verifies if the anonymization should be performed and returns a pointer to GRPCAnonymizer.
func checkAnonymizationInitializeGRPC(performAnonymizationBool bool) *GRPCAnonymizer {
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
