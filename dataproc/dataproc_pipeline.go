package dataproc

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/Kaszanas/GoSC2Science/utils"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func PipelineWrapper(absolutePathOutputDirectory string,
	chunks [][]string,
	integrityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	noMultiprocessing bool) {

	log.Info("Entered PipelineWrapper()")

	if noMultiprocessing {
		runtime.GOMAXPROCS(1)
	}

	for index, chunk := range chunks {
		go MultiprocessingChunkPipeline(absolutePathOutputDirectory,
			chunk,
			integrityCheckBool,
			gameModeCheckFlag,
			performAnonymizationBool,
			performCleanupBool,
			localizeMapsBool,
			localizedMapsMap,
			compressionMethod,
			index)
	}

	log.Info("Finished PipelineWrapper()")
}

func MultiprocessingChunkPipeline(absolutePathOutputDirectory string,
	listOfFiles []string,
	integrityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{},
	compressionMethod uint16,
	chunkIndex int) {

	log.Info("Entered MultiprocessingChunkPipeline()")

	// TODO: Create logging file:
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
			didWork, replayString, replaySummary := FileProcessingPipeline(replayFile,
				integrityCheckBool,
				gameModeCheckFlag,
				performAnonymizationBool,
				performCleanupBool,
				localizeMapsBool,
				localizedMapsMap)

			if !didWork {
				readErrorCounter++
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

			// Writing the zip archive to drive:
			writer.Close()
			packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(chunkIndex)+".zip")
			_ = ioutil.WriteFile(packageAbsPath, buffer.Bytes(), 0777)

			processedCounter++
			processingInfoStruct.ProcessedFiles = append(processingInfoStruct.ProcessedFiles, replayFile)

			// Saving contents of the persistent player nickname map and additional information about which package was processed:
			utils.SaveProcessingInfo(*processingInfoFile, processingInfoStruct)
			log.Info("Saved processing.log")
		}
	}

	// TODO: Write packageSummary to drive!!!

	// TODO: Save the ZIP archive:

	// Logging error information:
	if readErrorCounter > 0 {
		log.WithField("readErrors", readErrorCounter).Info("Finished processing ", readErrorCounter)
	}
	if compressionErrorCounter > 0 {
		log.WithField("compressionErrors", compressionErrorCounter).Info("Finished processing compressionErrors: ", compressionErrorCounter)
	}

	log.Info("Finished MultiprocessingChunkPipeline()")

}

// Pipeline is performing the whole data processing pipeline for a replay file. Reads the replay, cleans the replay structure, creates replay summary, anonymizes, and creates a JSON replay output.
func FileProcessingPipeline(replayFile string,
	integrityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{}) (bool, string, data.ReplaySummary) {

	log.Info("Entered FileProcessingPipeline()")

	// Read replay:
	replayData, err := rep.NewFromFile(replayFile)
	if err != nil {
		log.WithFields(log.Fields{"file": replayFile, "error": err, "readError": true}).Error("Failed to read file.")
		return false, "", data.ReplaySummary{}
	}
	log.WithField("file", replayFile).Info("Read data from a replay.")

	// Performing integrity checks
	integrityOk := checkIntegrity(replayData, integrityCheckBool, gameModeCheckFlag)
	if !integrityOk {
		log.WithField("file", replayData).Error("Integrity check failed in file.")
		if integrityCheckBool {
			return false, "", data.ReplaySummary{}
		}
	}

	// Clean replay structure:
	cleanOk, cleanReplayStructure := cleanReplay(replayData, localizeMapsBool, localizedMapsMap, performCleanupBool)
	if !cleanOk {
		log.WithField("file", replayFile).Error("Failed to perform cleaning.")
		return false, "", data.ReplaySummary{}
	}

	// Create replay summary:
	summarizeOk, summarizedReplay := summarizeReplay(&cleanReplayStructure)
	if !summarizeOk {
		log.WithField("file", replayFile).Error("Failed to create replay summary.")
		return false, "", data.ReplaySummary{}
	}

	// Anonimize replay:
	if !performAnonymizationBool {
		log.Info("Detected bypassAnonymizationBool, performing anonymization.")
		if !anonymizeReplay(&cleanReplayStructure) {
			log.WithField("file", replayFile).Error("Failed to anonymize replay.")
			return false, "", data.ReplaySummary{}
		}
	}

	// Create final replay string:
	stringifyOk, finalReplayString := stringifyReplay(&cleanReplayStructure)
	if !stringifyOk {
		log.WithField("file", replayFile).Error("Failed to stringify the replay.")
		return false, "", data.ReplaySummary{}
	}

	replayData.Close()
	log.Info("Closed replayData")

	log.Info("Finished FileProcessingPipeline()")

	return true, finalReplayString, summarizedReplay
}
