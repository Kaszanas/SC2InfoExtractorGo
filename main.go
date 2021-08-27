package main

import (
	"math"
	"os"
	"runtime/pprof"

	"archive/zip"
	"flag"
	"io"
	"path/filepath"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/larzconwell/bzip2"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Command line arguments:
	inputDirectory := flag.String("input", "./DEMOS/Input", "Input directory where .SC2Replay files are held.")
	outputDirectory := flag.String("output", "./DEMOS/Output", "Output directory where compressed zip packages will be saved.")
	numberOfPackagesFlag := flag.Int("number_of_packages", 1, "Provide a number of packages to be created and compressed into a zip archive. Please remember that this number needs to be lower than the number of processed files.")

	performIntegrityCheckFlag := flag.Bool("integrity_check", true, "If the software is supposed to check the hardcoded integrity checks for the provided replays")
	performValidityCheckFlag := flag.Bool("validity_check", true, "Provide if the tool is supposed to use hardcoded validity checks and verify if the replay file variables are within 'common sense' ranges.")

	// TODO: Write the docs for other game modes:
	gameModeCheckFlag := flag.Int("game_mode", 0b1111111111, "Provide which game mode should be included from the processed files in a format of a binary flag: AllGameModes: 0xFFFFFFFF")

	// Other compression methods than Deflate need to be registered further down in the code:
	compressionMethodFlag := flag.Int("compression_method", 8, "Provide a compression method number, default is 8 'Deflate', other compression methods need to be registered manually in code.")
	localizeMapsBoolFlag := flag.Bool("localize_maps", true, "Set to false if You want to keep the original (possibly foreign) map names.")
	localizationMappingFileFlag := flag.String("localized_maps_file", "./operation_files/output.json", "Specify a path to localization file containing {'ForeignName': 'EnglishName'} of maps.")

	performCleanupFlag := flag.Bool("perform_cleanup", true, "Provide if the tool is supposed to perform the cleaning functions within the processing pipeline.")
	performAnonymizationFlag := flag.Bool("perform_anonymization", true, "Provide if the tool is supposed to perform the anonymization functions within the processing pipeline.")

	processWithMultiprocessingFlag := flag.Bool("with_multiprocessing", true, "Provide if the processing is supposed to be perform with maximum amount of available cores. If set to false, the program will use one core.")

	logLevelFlag := flag.Int("log_level", 4, "Provide a log level from 1-7. Panic - 1, Fatal - 2, Error - 3, Warn - 4, Info - 5, Debug - 6, Trace - 7")
	performCPUProfilingFlag := flag.String("with_cpu_profiler", "", "Set path to the file where pprof cpu profiler will save its information. If this is empty no profiling is performed.")

	logDirectoryFlag := flag.String("log_dir", "./logs/", "Provide directory which will hold the logging information.")

	flag.Parse()
	logDirectoryString := *logDirectoryFlag
	log.SetFormatter(&log.JSONFormatter{})

	// If the file doesn't exist, create it or append to the file
	logFileFilepath := logDirectoryString + "main_log.log"
	logFile, err := os.OpenFile(logFileFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	log.WithField("logLevel", *logLevelFlag).Info("Parsed flags, setting log level.")
	log.SetLevel(log.Level(*logLevelFlag))
	log.Info("Set logging level.")

	performCPUProfilingPath := *performCPUProfilingFlag
	if performCPUProfilingPath != "" {
		// Creating profiler file:
		profilerFile, err := os.Create(performCPUProfilingPath)
		if err != nil {
			log.WithField("error", err).Fatal("Could not create a profiling file. Exiting program.")
			os.Exit(1)
		}
		// Starting profiling:
		pprof.StartCPUProfile(profilerFile)
		defer pprof.StopCPUProfile()
	}

	// Converting compression method flag:
	compressionMethod := uint16(*compressionMethodFlag)

	// Getting absolute path to input directory:
	absolutePathInputDirectory, _ := filepath.Abs(*inputDirectory)
	// absolutePathInterDirectory, _ := filepath.Abs(*interDirectory)
	absolutePathOutputDirectory, _ := filepath.Abs(*outputDirectory)

	performIntegrityCheckBool := *performIntegrityCheckFlag
	performValidityCheckBool := *performValidityCheckFlag

	// Filter game modes:
	filterGameModeFlag := *gameModeCheckFlag

	// Localization flags dereference:
	localizeMapsBool := *localizeMapsBoolFlag
	localizationMappingJSONFile := *localizationMappingFileFlag

	performAnonymizationBool := *performAnonymizationFlag
	performCleanupBool := *performCleanupFlag
	processWithMultiprocessingBool := *processWithMultiprocessingFlag

	numberOfPackages := *numberOfPackagesFlag

	log.WithFields(log.Fields{
		"inputDirectory":                 absolutePathInputDirectory,
		"outputDirectory":                absolutePathOutputDirectory,
		"numberOfPackages":               numberOfPackages,
		"compressionMethod":              compressionMethod,
		"filterGameModeFlag":             filterGameModeFlag,
		"localizeMapsBool":               localizeMapsBool,
		"localizationMappingJSONFile":    localizationMappingJSONFile,
		"performAnonymizationBool":       performAnonymizationBool,
		"performCleanupBool":             performCleanupBool,
		"processWithMultiprocessingBool": processWithMultiprocessingBool}).Info("Parsed command line flags")

	// Getting list of absolute paths for files from input directory:
	listOfInputFiles := utils.ListFiles(absolutePathInputDirectory, ".SC2Replay")
	lenListOfInputFiles := len(listOfInputFiles)
	if lenListOfInputFiles < numberOfPackages {
		log.WithFields(log.Fields{
			"lenListOfInputFiles": lenListOfInputFiles,
			"numberOfPackages":    numberOfPackages}).Error("Higher number of packages than input files, closing the program.")
		os.Exit(1)
	}
	numberOfFilesInPackage := int(math.Ceil(float64(lenListOfInputFiles) / float64(numberOfPackages)))
	listOfChunksFiles := chunkSlice(listOfInputFiles, numberOfFilesInPackage)

	// Register a custom compressor:
	zip.RegisterCompressor(12, func(out io.Writer) (io.WriteCloser, error) {
		return bzip2.NewWriterLevel(out, 9)
	})

	// Opening and marshalling the JSON to map[string]string to use in the pipeline (localization information of maps that were played).
	localizedMapsMap := utils.UnmarshalLocaleMapping(localizationMappingJSONFile)
	if localizedMapsMap == nil {
		log.Error("Could not read the JSON mapping file, closing the program.")
		os.Exit(1)
	}

	dataproc.PipelineWrapper(absolutePathOutputDirectory,
		listOfChunksFiles,
		performIntegrityCheckBool,
		performValidityCheckBool,
		filterGameModeFlag,
		performAnonymizationBool,
		performCleanupBool,
		localizeMapsBool,
		localizedMapsMap,
		compressionMethod,
		processWithMultiprocessingBool,
		logDirectoryString)

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
