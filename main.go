package main

import (
	"math"
	"os"
	"runtime/pprof"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Getting the information from user to start the processing:
	flags, okFlags := parseFlags()
	if !okFlags {
		log.Fatal("Failed parseFlags()")
		os.Exit(1)
	}

	logFile, okLogging := setLogging(flags.LogPath, flags.LogLevel)
	if !okLogging {
		log.Fatal("Failed to setLogging()")
		os.Exit(1)
	}

	if flags.CPUProfilingPath != "" {
		okProfiling := setProfiling(flags.CPUProfilingPath)
		if !okProfiling {
			log.Fatal("Failed to setProfiling()")
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	log.WithFields(log.Fields{
		"InputDirectory":        flags.InputDirectory,
		"OutputDirectory":       flags.OutputDirectory,
		"NumberOfPackages":      flags.NumberOfPackages,
		"PerformIntegrityCheck": flags.PerformIntegrityCheck,
		"PerformValidityCheck":  flags.PerformValidityCheck,
		"PerformCleanup":        flags.PerformCleanup,
		"PerformAnonymization":  flags.PerformAnonymization,
		"FilterGameMode":        flags.FilterGameMode,
		"LocalizationMapFile":   flags.LocalizationMapFile,
		"WithMultiprocessing":   flags.WithMultiprocessing,
		"LogLevel":              flags.LogLevel,
		"CPUProfilingPath":      flags.CPUProfilingPath,
		"LogPath":               flags.LogPath}).Info("Parsed command line flags")

	// Getting list of absolute paths for files from input directory:
	listOfInputFiles := utils.ListFiles(flags.InputDirectory, ".SC2Replay")
	lenListOfInputFiles := len(listOfInputFiles)
	if lenListOfInputFiles < flags.NumberOfPackages {
		log.WithFields(log.Fields{
			"lenListOfInputFiles": lenListOfInputFiles,
			"numberOfPackages":    flags.NumberOfPackages}).Error("Higher number of packages than input files, closing the program.")
		os.Exit(1)
	}
	numberOfFilesInPackage := int(math.Ceil(float64(lenListOfInputFiles) / float64(flags.NumberOfPackages)))
	listOfChunksFiles := chunkSlice(listOfInputFiles, numberOfFilesInPackage)

	// Opening and marshalling the JSON to map[string]string to use in the pipeline (localization information of maps that were played).
	localizedMapsMap := map[string]interface{}(nil)
	if flags.LocalizationMapFile != "" {
		localizedMapsMap := utils.UnmarshalLocaleMapping(flags.LocalizationMapFile)
		if localizedMapsMap == nil {
			log.Error("Could not read the JSON mapping file, closing the program.")
			os.Exit(1)
		}
	}

	dataproc.PipelineWrapper(flags.OutputDirectory,
		listOfChunksFiles,
		flags.PerformIntegrityCheck,
		flags.PerformValidityCheck,
		flags.FilterGameMode,
		flags.PerformAnonymization,
		flags.PerformCleanup,
		localizedMapsMap,
		8,
		flags.WithMultiprocessing,
		flags.LogPath)

	// Closing the log file:
	logFile.Close()
}

func chunkSlice(slice []string, chunkSize int) [][]string {

	log.Info("Entered chunkSlice()")

	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond slice capacity:
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	log.Info("Finished chunkSlice(), returning")
	return chunks
}
