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
	logFile, okLogging := utils.SetLogging(flags.LogPath, flags.LogLevel)
	if !okLogging {
		log.Fatal("Failed to setLogging()")
		return 1
	}

	log.WithFields(log.Fields{
		"InputDirectory":             flags.InputDirectory,
		"OutputDirectory":            flags.OutputDirectory,
		"NumberOfPackages":           flags.NumberOfPackages,
		"PerformIntegrityCheck":      flags.PerformIntegrityCheck,
		"PerformValidityCheck":       flags.PerformValidityCheck,
		"PerformCleanup":             flags.PerformCleanup,
		"PerformPlayerAnonymization": flags.PerformPlayerAnonymization,
		"PerformChatAnonymization":   flags.PerformChatAnonymization,
		"FilterGameMode":             flags.FilterGameMode,
		"LocalizationMapFile":        flags.LocalizationMapFile,
		"NumberOfThreads":            flags.NumberOfThreads,
		"LogLevel":                   flags.LogLevel,
		"CPUProfilingPath":           flags.CPUProfilingPath,
		"LogPath":                    flags.LogPath}).Info("Parsed command line flags")

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
			"lenListOfInputFiles": lenListOfInputFiles,
			"numberOfPackages":    flags.NumberOfPackages}).Error("Higher number of packages than input files, closing the program.")
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

	// TODO: Pass CLI Flags directly, limit the amount of arguments passed to the function:
	// Initializing the processing:
	dataproc.PipelineWrapper(
		flags.OutputDirectory,
		listOfChunksFiles,
		packageToZipBool,
		flags.PerformIntegrityCheck,
		flags.PerformValidityCheck,
		flags.PerformFiltering,
		flags.FilterGameMode,
		flags.PerformPlayerAnonymization,
		flags.PerformChatAnonymization,
		flags.PerformCleanup,
		localizedMapsMap,
		8,
		flags.NumberOfThreads,
		flags.LogPath)

	// Closing the log file manually:
	logFile.Close()

	return 0
}
