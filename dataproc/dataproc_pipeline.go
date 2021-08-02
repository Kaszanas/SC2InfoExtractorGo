package dataproc

import (
	"runtime"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

func PipelineWrapper(chunks [][]string,
	integrityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{},
	noMultiprocessing bool) {

	if noMultiprocessing {
		runtime.GOMAXPROCS(1)
	}

	for index, chunk := range chunks {
		go MultiprocessingChunkPipeline(chunk,
			integrityCheckBool,
			gameModeCheckFlag,
			performAnonymizationBool,
			performCleanupBool,
			localizeMapsBool,
			localizedMapsMap,
			index)
	}

}

func MultiprocessingChunkPipeline(listOfFiles []string,
	integrityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{},
	chunkIndex int) {

	// TODO: Create logging file:

	// Defining counters:
	readErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0

	// Helper method returning bytes buffer and zip writer:
	buffer, writer := initBufferWriter()
	log.Info("Initialized buffer and writer.")

	// Processing file:
	for _, file := range listOfFiles {

		didWork, replayString, replaySummary := FileProcessingPipeline(file,
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

	}

	// TODO: Save the ZIP archive:

}

// Pipeline is performing the whole data processing pipeline for a replay file. Reads the replay, cleans the replay structure, creates replay summary, anonymizes, and creates a JSON replay output.
func FileProcessingPipeline(replayFile string,
	integrityCheckBool bool,
	gameModeCheckFlag int,
	performAnonymizationBool bool,
	performCleanupBool bool,
	localizeMapsBool bool,
	localizedMapsMap map[string]interface{}) (bool, string, data.ReplaySummary) {

	log.Info("Entered Pipeline()")

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

	log.Info("Finished Pipeline()")
	return true, finalReplayString, summarizedReplay
}
