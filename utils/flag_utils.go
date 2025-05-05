package utils

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	log "github.com/sirupsen/logrus"
)

// LogFlags contains settings that the user can set for logging.
type LogFlags struct {
	LogLevelValue datastruct.LogLevelEnum
	LogPath       string
}

// CLIFlags is a structure which holds all of the information that was supplied by user in CLI.
type CLIFlags struct {
	InputDirectory             string
	OutputDirectory            string
	OnlyDependencyDownload     bool
	SkipDependencyDownload     bool
	DependencyDirectory        string
	NumberOfThreads            int
	NumberOfPackages           int
	PerformIntegrityCheck      bool
	PerformValidityCheck       bool
	PerformCleanup             bool
	PerformPlayerAnonymization bool
	PerformChatAnonymization   bool
	PerformFiltering           bool
	FilterGameMode             int
	LogFlags                   LogFlags
	CPUProfilingPath           string
}

// ParseFlags contains logic which is responsible for user input.
func ParseFlags() (CLIFlags, bool) {
	// Command line arguments:
	inputDirectory := flag.String(
		"input",
		"./replays/input",
		"Input directory where .SC2Replay files are held.",
	)
	outputDirectory := flag.String(
		"output",
		"./replays/output",
		"Output directory where compressed zip packages will be saved.",
	)

	onlyDependencyDownload := flag.Bool(
		"only_dependency_download",
		false,
		`Flag specifying if the tool is supposed to only download
		the replay dependencies and not process the replays.`,
	)
	skipDependencyDownload := flag.Bool(
		"skip_dependency_download",
		false,
		`Flag specifying if the tool is supposed to skip the dependency download.`,
	)

	dependencyDirectory := flag.String(
		"dependency_directory",
		"./dependencies/",
		"Directory where the replay dependencies will be downloaded as a result of the replay processing.",
	)

	numberOfPackagesFlag := flag.Int(
		"number_of_packages",
		1,
		`Provide a number of zip packages to be created and compressed
		into a zip archive. Please remember that this number needs to be lower
		than the number of processed files. If set to 0, will ommit the
		zip packaging and output .json directly to drive.`,
	)

	// Boolean Flags:
	help := flag.Bool(
		"help",
		false,
		"Show command usage")
	performIntegrityCheckFlag := flag.Bool(
		"perform_integrity_checks",
		false,
		`Flag specifying if the software is supposed to check the hardcoded
		integrity checks for the provided replays`,
	)
	performValidityCheckFlag := flag.Bool(
		"perform_validity_checks",
		false,
		`Flag, specifying if the tool is supposed to use hardcoded validity checks
		and verify if the replay file variables are within 'common sense' ranges.`,
	)
	performCleanupFlag := flag.Bool(
		"perform_cleanup",
		false,
		`Flag specifying if the tool is supposed to perform the cleaning
		functions within the processing pipeline.`,
	)
	performPlayerAnonymizationFlag := flag.Bool(
		"perform_player_anonymization",
		false,
		`Flag specifying if the tool is supposed to perform player anonymization
		functions within the processing pipeline.
		If set to true please remember to download and run
		an anonymization server: https://doi.org/10.5281/zenodo.5138313`,
	)
	performChatAnonymizationFlag := flag.Bool(
		"perform_chat_anonymization",
		false,
		"Flag, specifying if the chat anonymization should be performed.",
	)

	// TODO: Write the docs for other game modes:
	performFilteringFlag := flag.Bool(
		"perform_filtering",
		false,
		`Flag, specifying if the pipeline ought to verify different hard coded game modes.
		If set to false completely bypasses the filtering.`,
	)
	gameModeFilterFlag := flag.Int(
		"game_mode_filter",
		0b11111111,
		`Specifies which game mode should be included from the processed files
		in a format of a binary flag: AllGameModes: 0b11111111 (default 0b11111111)`,
	)

	// processWithMultiprocessingFlag := flag.Bool("with_multiprocessing", false, "Specifies if the processing is supposed to be perform with maximum amount of available cores. If set to false, the program will use one core.")
	numberOfThreadsUsedFlag := flag.Int(
		"max_procs",
		runtime.NumCPU(),
		"Specifies the number of logic cores of a processor that will be used for processing (default runtime.NumCPU()).",
	)

	// Misc flags:
	logLevelFlag := flag.Int(
		"log_level",
		4,
		`Specifies a log level from 1-7:
		Panic - 1, Fatal - 2,
		Error - 3, Warn - 4,
		Info - 5, Debug - 6,
		Trace - 7`,
	)
	logDirectoryFlag := flag.String(
		"log_dir",
		"./logs/",
		"Specifies directory which will hold the logging information.",
	)
	performCPUProfilingFlag := flag.String(
		"with_cpu_profiler",
		"",
		`Set path to the file where pprof cpu profiler will save its information.
		If this is empty no profiling is performed.`,
	)

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	absoluteInputDirectory, err := filepath.Abs(*inputDirectory)
	if err != nil {
		log.WithField("inputDirectory", *inputDirectory).
			Error("Failed to get the absolute path to the input directory!")
		return CLIFlags{}, false
	}

	absolutePathOutputDirectory, err := filepath.Abs(*outputDirectory)
	if err != nil {
		log.WithField("outputDirectory", *outputDirectory).
			Error("Failed to get the absolute path to the output directory!")
		return CLIFlags{}, false
	}

	absolutePathDependencyDirectory, err := filepath.Abs(*dependencyDirectory)
	if err != nil {
		log.WithField("dependencyDirectory", *dependencyDirectory).
			Error("Failed to get the absolute path to the dependency directory!")
		return CLIFlags{}, false
	}

	logFlags := LogFlags{
		LogLevelValue: datastruct.LogLevelEnum(*logLevelFlag),
		LogPath:       *logDirectoryFlag,
	}

	flags := CLIFlags{
		InputDirectory:             absoluteInputDirectory,
		OutputDirectory:            absolutePathOutputDirectory,
		OnlyDependencyDownload:     *onlyDependencyDownload,
		SkipDependencyDownload:     *skipDependencyDownload,
		DependencyDirectory:        absolutePathDependencyDirectory,
		NumberOfPackages:           *numberOfPackagesFlag,
		PerformIntegrityCheck:      *performIntegrityCheckFlag,
		PerformValidityCheck:       *performValidityCheckFlag,
		PerformCleanup:             *performCleanupFlag,
		PerformPlayerAnonymization: *performPlayerAnonymizationFlag,
		PerformChatAnonymization:   *performChatAnonymizationFlag,
		PerformFiltering:           *performFilteringFlag,
		FilterGameMode:             *gameModeFilterFlag,
		NumberOfThreads:            *numberOfThreadsUsedFlag,
		LogFlags:                   logFlags,
		CPUProfilingPath:           *performCPUProfilingFlag,
	}

	return flags, true
}
