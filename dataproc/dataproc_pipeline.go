package dataproc

import (
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

// PipelineWrapper is an orchestrator that distributes work among available workers (threads)
func PipelineWrapper(absolutePathOutputDirectory string,
	chunks [][]string,
	performIntegrityCheckBool bool,
	performValidityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	withMultiprocessing bool,
	logsFilepath string) {

	log.Info("Entered PipelineWrapper()")

	// If it is specified by the user to perform the processing without multiprocessing GOMACPROCS needs to be set to 1 in order to allow 1 thread:
	if !withMultiprocessing {
		runtime.GOMAXPROCS(1)
	}

	var wg sync.WaitGroup

	// Adding a task for each of the supplied chunks to speed up the processing:
	for index, chunk := range chunks {
		wg.Add(1)
		go MultiprocessingChunkPipeline(absolutePathOutputDirectory,
			chunk,
			performIntegrityCheckBool,
			performValidityCheckBool,
			gameModeCheckFlag,
			performAnonymizationBool,
			performCleanupBool,
			localizedMapsMap,
			compressionMethod,
			logsFilepath,
			index,
			&wg)
	}
	wg.Wait()

	log.Info("Finished PipelineWrapper()")
}

// MultiprocessingChunkPipeline is a single instance of processing that is meant to be spawned by the orchestrator in order to speed up the process of data extraction.
func MultiprocessingChunkPipeline(absolutePathOutputDirectory string,
	listOfFiles []string,
	performIntegrityCheckBool bool,
	performValidityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	logsFilepath string,
	chunkIndex int,
	waitGroup *sync.WaitGroup) {

	// Letting the orchestrator know that this processing task was finished:
	defer waitGroup.Done()
	log.Info("Entered MultiprocessingChunkPipeline()")

	// Create ProcessingInfoFile:
	processingInfoFile, processingInfoStruct := utils.CreateProcessingInfoFile(logsFilepath, chunkIndex)
	defer processingInfoFile.Close()

	// Initializing grpc connection if the user chose to perform anonymization.
	var grpcAnonymizer *GRPCAnonymizer
	if performAnonymizationBool {
		grpcAnonymizer := GRPCAnonymizer{}
		if !grpcAnonymizer.grpcConnect() {
			log.Error("Could not connect to the gRPC server!")
		}

		defer grpcAnonymizer.Connection.Close()

	}

	// Defining counters:
	readErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0

	// Helper method returning bytes buffer and zip writer which will be used to save the processing results into:
	buffer, writer := utils.InitBufferWriter()
	log.Info("Initialized buffer and writer.")

	// Create package summary structure:
	packageSummary := data.DefaultPackageSummary()
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
			performCleanupBool,
			localizedMapsMap)

		if !didWork {
			readErrorCounter++
			processingInfoStruct.FailedToProcess = append(processingInfoStruct.FailedToProcess, map[string]string{replayFile: failureReason})
			continue
		}

		// Append it to a list and when a package is created create a package summary and clear the list for next iterations
		data.AddReplaySummToPackageSumm(&replaySummary, &packageSummary)
		log.Info("Added replaySummary to packageSummary")

		// Saving output to zip archive:
		savedSuccess := utils.SaveFileToArchive(replayString, replayFile, compressionMethod, writer)
		if !savedSuccess {
			compressionErrorCounter++
			continue
		}
		log.Info("Added file to zip archive.")

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

	// Logging error information:
	if readErrorCounter > 0 {
		log.WithField("readErrors", readErrorCounter).Info("Finished processing ", readErrorCounter)
	}
	if compressionErrorCounter > 0 {
		log.WithField("compressionErrors", compressionErrorCounter).Info("Finished processing compressionErrors: ", compressionErrorCounter)
	}

	log.Info("Finished MultiprocessingChunkPipeline()")

}

// FileProcessingPipeline is performing the whole data processing pipeline for a replay file. Reads the replay, cleans the replay structure, creates replay summary, anonymizes, and creates a JSON replay output.
func FileProcessingPipeline(replayFile string,
	performIntegrityCheckBool bool,
	performValidityCheckBool bool,
	gameModeCheckFlag int,
	grpcAnonymizer *GRPCAnonymizer,
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
		return false, "", data.ReplaySummary{}, "checkGameMode() failed"
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
	if grpcAnonymizer != nil && !anonymizeReplay(&cleanReplayStructure, grpcAnonymizer) {
		log.WithField("file", replayFile).Error("Failed to anonymize replay.")
		return false, "", data.ReplaySummary{}, "anonymizeReplay() failed"
	}

	// Create final replay string:
	stringifyOk, finalReplayString := stringifyReplay(&cleanReplayStructure)
	if !stringifyOk {
		log.WithField("file", replayFile).Error("Failed to stringify the replay.")
		return false, "", data.ReplaySummary{}, "stringifyReplay() failed"
	}

	replayData.Close()
	log.Info("Closed replayData")

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
