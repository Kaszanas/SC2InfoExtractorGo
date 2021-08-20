package dataproc

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/Kaszanas/GoSC2Science/utils"
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
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	withMultiprocessing bool) {

	log.Info("Entered PipelineWrapper()")

	if !withMultiprocessing {
		runtime.GOMAXPROCS(1)
	}

	var wg sync.WaitGroup

	for index, chunk := range chunks {
		wg.Add(1)
		go MultiprocessingChunkPipeline(absolutePathOutputDirectory,
			chunk,
			performIntegrityCheckBool,
			performValidityCheckBool,
			gameModeCheckFlag,
			performAnonymizationBool,
			performCleanupBool,
			localizeMapsBool,
			localizedMapsMap,
			compressionMethod,
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
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	chunkIndex int,
	waitGroup *sync.WaitGroup) {

	// Letting the orchestrator know that this processing task was finished:
	defer waitGroup.Done()
	log.Info("Entered MultiprocessingChunkPipeline()")

	// Create ProcessingInfoFile:
	processingInfoFile, processingInfoStruct := utils.CreateProcessingInfoFile(chunkIndex)
	defer processingInfoFile.Close()

	// Defining counters:
	readErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0

	// Helper method returning bytes buffer and zip writer:
	buffer, writer := utils.InitBufferWriter()
	log.Info("Initialized buffer and writer.")

	// Create package summary structure:
	packageSummary := data.DefaultPackageSummary()
	// Processing file:
	for _, replayFile := range listOfFiles {
		// Checking if the file was previously processed:
		if !contains(processingInfoStruct.ProcessedFiles, replayFile) {
			didWork, replayString, replaySummary, failureReason := FileProcessingPipeline(replayFile,
				performIntegrityCheckBool,
				performValidityCheckBool,
				gameModeCheckFlag,
				performAnonymizationBool,
				performCleanupBool,
				localizeMapsBool,
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

	}

	// Saving processingInfo to know which files failed to process:
	utils.SaveProcessingInfo(processingInfoFile, processingInfoStruct)
	log.Info("Saved processing.log")

	// Writing PackageSummaryFile to drive:
	utils.CreatePackageSummaryFile(absolutePathOutputDirectory, packageSummary, chunkIndex)

	// Writing the zip archive to drive:
	writer.Close()
	packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(chunkIndex)+".zip")
	_ = ioutil.WriteFile(packageAbsPath, buffer.Bytes(), 0777)

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
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
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
	if performIntegrityCheckBool {
		integrityOk := checkIntegrity(replayData)
		if !integrityOk {
			log.WithField("file", replayFile).Error("Integrity check failed in file.")
			if performIntegrityCheckBool {
				return false, "", data.ReplaySummary{}, "checkIntegrity() failed"
			}
		}
	}

	// Performing validity checks:
	// TODO: Add validity check flag to the CLI
	if performValidityCheckBool {
		if gameModeCheckFlag&Ranked1v1 != 0 && gameIs1v1Ranked(replayData) {
			// Perform Validity check
			if !validateReplay(replayData) {
				return false, "", data.ReplaySummary{}, "validateReplay() failed"
			}
		}
	}

	// Filtering
	if !checkGameMode(replayData, gameModeCheckFlag) {
		return false, "", data.ReplaySummary{}, "checkGameMode() failed"
	}

	// Clean replay structure:
	cleanOk, cleanReplayStructure := cleanReplay(replayData, localizeMapsBool, localizedMapsMap, performCleanupBool)
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
	if performAnonymizationBool {
		log.Info("Detected bypassAnonymizationBool, performing anonymization.")
		if !anonymizeReplay(&cleanReplayStructure) {
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
