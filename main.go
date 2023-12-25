package main

import (
	"os"
	"runtime/pprof"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func main() {
	os.Exit(mainReturnWithCode())
}

// TODO: Wrap the main functionality do not call os.exit directly it does not run deferred functions.
func mainReturnWithCode() int {

	// Getting the information from user to start the processing:
	flags, okFlags := utils.ParseFlags()
	if !okFlags {
		log.Fatal("Failed parseFlags()")
		return 1
	}

	// Logging initialization to be able to provide further troubleshooting for users:
	logFile, okLogging := utils.SetLogging(flags.LogFlags.LogPath, flags.LogFlags.LogLevel)
	if !okLogging {
		log.Fatal("Failed to setLogging()")
		return 1
	}

	log.WithFields(log.Fields{
		"flags.InputDirectory":             flags.InputDirectory,
		"flags.OutputDirectory":            flags.OutputDirectory,
		"flags.NumberOfPackages":           flags.NumberOfPackages,
		"flags.PerformIntegrityCheck":      flags.PerformIntegrityCheck,
		"flags.PerformValidityCheck":       flags.PerformValidityCheck,
		"flags.PerformCleanup":             flags.PerformCleanup,
		"flags.PerformPlayerAnonymization": flags.PerformPlayerAnonymization,
		"flags.PerformChatAnonymization":   flags.PerformChatAnonymization,
		"flags.FilterGameMode":             flags.FilterGameMode,
		"flags.LocalizationMapFile":        flags.LocalizationMapFile,
		"flags.NumberOfThreads":            flags.NumberOfThreads,
		"flags.LogFlags.LogLevel":          flags.LogFlags.LogLevel,
		"flags.LogFlags.LogPath":           flags.LogFlags.LogPath,
		"flags.CPUProfilingPath":           flags.CPUProfilingPath,
	}).Info("Parsed command line flags")

	// Profiling capabilities to verify if the program can be optimized any further:
	if flags.CPUProfilingPath != "" {
		_, okProfiling := utils.SetProfiling(flags.CPUProfilingPath)
		if !okProfiling {
			log.Fatal("Failed to setProfiling()")
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	// TODO: Move everything that is below to separate functions:
	// Getting list of absolute paths for files from input directory filtering them by file extension to be able to extract the data:
	listOfInputFiles := utils.ListFiles(flags.InputDirectory, ".SC2Replay")
	lenListOfInputFiles := len(listOfInputFiles)
	if lenListOfInputFiles < flags.NumberOfPackages {
		log.WithFields(log.Fields{
			"lenListOfInputFiles":    lenListOfInputFiles,
			"flags.NumberOfPackages": flags.NumberOfPackages}).Error("Higher number of packages than input files, closing the program.")
		return 1
	}

	listOfChunksFiles, packageToZipBool := utils.GetChunkListAndPackageBool(
		listOfInputFiles,
		flags.NumberOfPackages,
		flags.NumberOfThreads,
		lenListOfInputFiles)

	// Opening and marshalling the JSON to map[string]string to use in the pipeline (localization information of maps that were played).
	localizedMapsMap := map[string]interface{}(nil)
	if flags.LocalizationMapFile != "" {
		localizedMapsMap = utils.UnmarshalLocaleMapping(flags.LocalizationMapFile)
		if localizedMapsMap == nil {
			log.Error("Could not read the JSON mapping file, closing the program.")
			return 1
		}
	}

	var compressionMethod uint16 = 8
	// TODO: Pass CLI Flags directly, limit the amount of arguments passed to the function:
	// Initializing the processing:
	dataproc.PipelineWrapper(
		listOfChunksFiles,
		packageToZipBool,
		localizedMapsMap,
		compressionMethod,
		flags,
	)

	// Closing the log file manually:
	logFile.Close()

	return 0
}
