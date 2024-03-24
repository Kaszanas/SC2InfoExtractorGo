package main

import (
	"os"
	"runtime/pprof"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func main() {
	// main function is wrapping mainReturnWith code as not to call os.Exit directly.
	// This is because os.Exit does not run deferred functions.
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {

	// Getting the information from user to start the processing:
	CLIflags, okFlags := utils.ParseFlags()
	if !okFlags {
		log.Fatal("Failed parseFlags()")
		return 1
	}

	// Logging initialization to be able to provide further troubleshooting for users:
	logFile, okLogging := utils.SetLogging(
		CLIflags.LogFlags.LogPath,
		CLIflags.LogFlags.LogLevel)
	if !okLogging {
		log.Fatal("Failed to setLogging()")
		return 1
	}

	log.WithFields(log.Fields{
		"CLIflags.InputDirectory":             CLIflags.InputDirectory,
		"CLIflags.OutputDirectory":            CLIflags.OutputDirectory,
		"CLIflags.NumberOfPackages":           CLIflags.NumberOfPackages,
		"CLIflags.PerformIntegrityCheck":      CLIflags.PerformIntegrityCheck,
		"CLIflags.PerformValidityCheck":       CLIflags.PerformValidityCheck,
		"CLIflags.PerformCleanup":             CLIflags.PerformCleanup,
		"CLIflags.PerformPlayerAnonymization": CLIflags.PerformPlayerAnonymization,
		"CLIflags.PerformChatAnonymization":   CLIflags.PerformChatAnonymization,
		"CLIflags.FilterGameMode":             CLIflags.FilterGameMode,
		"CLIflags.LocalizationMapFile":        CLIflags.LocalizationMapFile,
		"CLIflags.NumberOfThreads":            CLIflags.NumberOfThreads,
		"CLIflags.LogFlags.LogLevel":          CLIflags.LogFlags.LogLevel,
		"CLIflags.LogFlags.LogPath":           CLIflags.LogFlags.LogPath,
		"CLIflags.CPUProfilingPath":           CLIflags.CPUProfilingPath,
	}).Info("Parsed command line flags")

	// Profiling capabilities to verify if the program can be optimized any further:
	if CLIflags.CPUProfilingPath != "" {
		_, okProfiling := utils.SetProfiling(CLIflags.CPUProfilingPath)
		if !okProfiling {
			log.Fatal("Failed to setProfiling()")
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	// TODO: Move everything that is below to separate functions:
	// Getting list of absolute paths for files from input
	// directory filtering them by file extension to be able to extract the data:
	listOfInputFiles := utils.ListFiles(CLIflags.InputDirectory, ".SC2Replay")
	lenListOfInputFiles := len(listOfInputFiles)
	if lenListOfInputFiles < CLIflags.NumberOfPackages {
		log.WithFields(log.Fields{
			"lenListOfInputFiles":    lenListOfInputFiles,
			"flags.NumberOfPackages": CLIflags.NumberOfPackages}).Error(
			"Higher number of packages than input files, closing the program.")
		return 1
	}

	listOfChunksFiles, packageToZipBool := utils.GetChunkListAndPackageBool(
		listOfInputFiles,
		CLIflags.NumberOfPackages,
		CLIflags.NumberOfThreads,
		lenListOfInputFiles)

	// Opening and marshalling the JSON to map[string]string to use in the pipeline (localization information of maps that were played).
	localizedMapsMap := map[string]interface{}(nil)
	if CLIflags.LocalizationMapFile != "" {
		localizedMapsMap = utils.UnmarshalLocaleMapping(CLIflags.LocalizationMapFile)
		if localizedMapsMap == nil {
			log.Error("Could not read the JSON mapping file, closing the program.")
			return 1
		}
	}

	var compressionMethod uint16 = 8
	// Initializing the processing:
	dataproc.PipelineWrapper(
		listOfChunksFiles,
		packageToZipBool,
		localizedMapsMap,
		compressionMethod,
		CLIflags,
	)

	// Closing the log file manually:
	logFile.Close()

	return 0
}
