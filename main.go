package main

import (
	"os"
	"runtime/pprof"

	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"archive/zip"
	"flag"
	"io"
	"path/filepath"

	"github.com/Kaszanas/GoSC2Science/dataproc"
	"github.com/Kaszanas/GoSC2Science/utils"
	"github.com/larzconwell/bzip2"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetFormatter(&log.JSONFormatter{})

	// If the file doesn't exist, create it or append to the file
	logFile, err := os.OpenFile("./logs/main_log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	// Command line arguments:
	inputDirectory := flag.String("input", "./DEMOS/Input", "Input directory where .SC2Replay files are held.")
	// interDirectory := flag.String("inter", "./Demos/Intermediate", "Intermediate directory where .json files will be stored before bzip2 compression.")
	outputDirectory := flag.String("output", "./DEMOS/Output", "Output directory where compressed bzip2 packages will be stored.")
	numberOfPackagesFlag := flag.Int("number_of_packaged_files", 100, "Provide a number of files to be packaged to be compressed into a zip archive. Please remember that this number need to be lower than the number of processed files.")

	performIntegrityCheckFlag := flag.Bool("integrity_check", true, "If the software is supposed to check the hardcoded integrity checks for the provided replays")
	performValidityCheckFlag := flag.Bool("validity_check", true, "Provide if the tool is supposed to use hardcoded validity checks and verify if the replay file variables are within 'common sense' ranges.")

	// TODO: Write the docs for other game modes:
	gameModeCheckFlag := flag.Int("game_mode", 0xFFFFFFFF, "Provide which game mode should be included from the processed files in a format of a binary flag: AllGameModes: 0xFFFFFFFF")

	// Other compression methods than Deflate need to be registered further down in the code:
	compressionMethodFlag := flag.Int("compression_method", 8, "Provide a compression method number, default is 8 'Deflate', other compression methods need to be registered in code.")
	localizeMapsBoolFlag := flag.Bool("localize_maps", true, "Set to false if You want to keep the original (possibly foreign) map names.")
	localizationMappingFileFlag := flag.String("localized_maps_file", "./operation_files/output.json", "Specify a path to localization file containing {'ForeignName': 'EnglishName'} of maps.")

	performCleanupFlag := flag.Bool("perform_cleanup", true, "Provide if the tool is supposed to bypass the cleaning functions within the processing pipeline.")
	performAnonymizationFlag := flag.Bool("perform_anonymization", true, "Provide if the tool is supposed to bypass the anonymization functions within the processing pipeline.")

	processWithMultiprocessingFlag := flag.Bool("with_multiprocessing", true, "Provide if the processing is supposed to be perform with maximum amount of available cores. If set to false, the program will use one core.")

	logLevelFlag := flag.Int("log_level", 4, "Provide a log level from 1-7. Panic - 1, Fatal - 2, Error - 3, Warn - 4, Info - 5, Debug - 6, Trace - 7")
	performCPUProfilingFlag := flag.String("with_cpu_profiler", "", "Set path to the file where pprof cpu profiler will save its information. If this is empty no profiling is performed.")

	flag.Parse()
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
		"inputDirectory":    absolutePathInputDirectory,
		"outputDirectory":   absolutePathOutputDirectory,
		"filesInPackage":    numberOfPackages,
		"compressionMethod": compressionMethod}).Info("Parsed command line flags")

	// Getting list of absolute paths for files from input directory:
	listOfInputFiles := utils.ListFiles(absolutePathInputDirectory, ".SC2Replay")
	listOfChunksFiles := chunkSlice(listOfInputFiles, numberOfPackages)

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
		processWithMultiprocessingBool)

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
