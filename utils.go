package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime/pprof"

	log "github.com/sirupsen/logrus"
)

// CLIFlags is a structure which holds all of the information that was supplied by user in CLI.
type CLIFlags struct {
	InputDirectory        string
	OutputDirectory       string
	NumberOfPackages      int
	PerformIntegrityCheck bool
	PerformValidityCheck  bool
	PerformCleanup        bool
	PerformAnonymization  bool
	FilterGameMode        int
	LocalizationMapFile   string
	WithMultiprocessing   bool
	LogLevel              int
	CPUProfilingPath      string
	LogPath               string
}

// parseFlags contains logic which is responsible for user input.
func parseFlags() (CLIFlags, bool) {
	// Command line arguments:
	inputDirectory := flag.String("input", "./DEMOS/Input", "Input directory where .SC2Replay files are held.")
	outputDirectory := flag.String("output", "./DEMOS/Output", "Output directory where compressed zip packages will be saved.")
	numberOfPackagesFlag := flag.Int("number_of_packages", 1, "Provide a number of zip packages to be created and compressed into a zip archive. Please remember that this number needs to be lower than the number of processed files.")

	// Boolean Flags:
	performIntegrityCheckFlag := flag.Bool("perform_integrity_checks", false, "If the software is supposed to check the hardcoded integrity checks for the provided replays")
	performValidityCheckFlag := flag.Bool("perform_validity_checks", false, "Provide if the tool is supposed to use hardcoded validity checks and verify if the replay file variables are within 'common sense' ranges.")
	performCleanupFlag := flag.Bool("perform_cleanup", false, "Provide if the tool is supposed to perform the cleaning functions within the processing pipeline.")
	performAnonymizationFlag := flag.Bool("perform_anonymization", false, "Provide if the tool is supposed to perform the anonymization functions within the processing pipeline. If set to true please remember to download and run an anonymization server: https://doi.org/10.5281/zenodo.5138313")

	// TODO: Write the docs for other game modes:
	gameModeFilterFlag := flag.Int("game_mode", 0b1111111111, "Provide which game mode should be included from the processed files in a format of a binary flag: AllGameModes: 0b1111111111")

	// Other compression methods than Deflate need to be registered further down in the code:
	localizationMappingFileFlag := flag.String("localized_maps_file", "./operation_files/output.json", "Specifies a path to localization file containing {'ForeignName': 'EnglishName'} of maps.")
	processWithMultiprocessingFlag := flag.Bool("with_multiprocessing", false, "Specifies if the processing is supposed to be perform with maximum amount of available cores. If set to false, the program will use one core.")

	// Misc flags:
	logLevelFlag := flag.Int("log_level", 4, "Specifies a log level from 1-7. Panic - 1, Fatal - 2, Error - 3, Warn - 4, Info - 5, Debug - 6, Trace - 7")
	logDirectoryFlag := flag.String("log_dir", "./logs/", "Specifies directory which will hold the logging information.")
	performCPUProfilingFlag := flag.String("with_cpu_profiler", "", "Set path to the file where pprof cpu profiler will save its information. If this is empty no profiling is performed.")

	flag.Parse()

	absoluteInputDirectory, err := filepath.Abs(*inputDirectory)
	if err != nil {
		log.WithField("inputDirectory", *inputDirectory).Error("Failed to get the absolute path to the input directory!")
		return CLIFlags{}, false
	}

	absolutePathOutputDirectory, err := filepath.Abs(*outputDirectory)
	if err != nil {
		log.WithField("outputDirectory", *outputDirectory).Error("Failed to get the absolute path to the output directory!")
		return CLIFlags{}, false
	}

	flags := CLIFlags{
		InputDirectory:        absoluteInputDirectory,
		OutputDirectory:       absolutePathOutputDirectory,
		NumberOfPackages:      *numberOfPackagesFlag,
		PerformIntegrityCheck: *performIntegrityCheckFlag,
		PerformValidityCheck:  *performValidityCheckFlag,
		PerformCleanup:        *performCleanupFlag,
		PerformAnonymization:  *performAnonymizationFlag,
		FilterGameMode:        *gameModeFilterFlag,
		LocalizationMapFile:   *localizationMappingFileFlag,
		WithMultiprocessing:   *processWithMultiprocessingFlag,
		LogLevel:              *logLevelFlag,
		CPUProfilingPath:      *performCPUProfilingFlag,
		LogPath:               *logDirectoryFlag,
	}

	// flag.Usage()

	return flags, true

}

// setLogging contains logic that is used to initialize logging to a specified file with a specified level.
func setLogging(logPath string, logLevel int) (*os.File, bool) {

	logDirectoryString := logPath
	log.SetFormatter(&log.JSONFormatter{})

	// If the file doesn't exist, create it or append to the file
	logFileFilepath := logDirectoryString + "main_log.log"
	logFile, err := os.OpenFile(logFileFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
		return &os.File{}, false
	}

	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	log.SetLevel(log.Level(logLevel))
	log.Info("Set logging level.")

	return logFile, true

}

// setProfiling sets up pprof profiling to a supplied filepath.
func setProfiling(profilingPath string) bool {

	performCPUProfilingPath := profilingPath

	// Creating profiler file:
	profilerFile, err := os.Create(performCPUProfilingPath)
	if err != nil {
		log.WithField("error", err).Error("Could not create a profiling file. Exiting program.")
		return false
	}
	// Starting profiling:
	pprof.StartCPUProfile(profilerFile)

	return true
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
