package main

import (
	"math"
	"os"
	"runtime/pprof"

	"archive/zip"
	"io"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/larzconwell/bzip2"
	log "github.com/sirupsen/logrus"
)

func main() {

	// Getting the information from user to start the processing:
	flags, okFlags := parseFlags()
	if !okFlags {
		log.Fatal("Failed parseFlags()")
	}

	logDirectoryString := flags.LogPath
	log.SetFormatter(&log.JSONFormatter{})

	// If the file doesn't exist, create it or append to the file
	logFileFilepath := logDirectoryString + "main_log.log"
	logFile, err := os.OpenFile(logFileFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	log.SetLevel(log.Level(flags.LogLevel))
	log.Info("Set logging level.")

	performCPUProfilingPath := flags.CPUProfilingPath
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

	// Register a custom compressor:
	zip.RegisterCompressor(12, func(out io.Writer) (io.WriteCloser, error) {
		return bzip2.NewWriterLevel(out, 9)
	})

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
