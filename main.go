package main

import (
	"fmt"
	"os"

	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"archive/zip"
	"flag"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/Kaszanas/GoSC2Science/dataproc"
	"github.com/larzconwell/bzip2"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.SetFormatter(&log.JSONFormatter{})

	// If the file doesn't exist, create it or append to the file
	logFile, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	log.Info("Set logging format, defined log file.")

	// Command line arguments:
	inputDirectory := flag.String("input", "./DEMOS/Input", "Input directory where .SC2Replay files are held.")
	// interDirectory := flag.String("inter", "./Demos/Intermediate", "Intermediate directory where .json files will be stored before bzip2 compression.")
	outputDirectory := flag.String("output", "./DEMOS/Output", "Output directory where compressed bzip2 packages will be stored.")
	filesInPackage := flag.Int("files_in_package", 3, "Provide a number of files to be compressed into a bzip2 archive.")
	// Other compression methods than Deflate need to be registered further down in the code:
	compressionMethodFlag := flag.Int("compression_method", 8, "Provide a compression method number, default is 8 'Deflate', other compression methods need to be registered in code.")

	logLevelFlag := flag.Int("log_level", 5, "Provide a log level from 1-7. Panic - 1, Fatal - 2, Error - 3, Warn - 4, Info - 5, Debug - 6, Trace - 7")

	flag.Parse()
	log.WithField("logLevel", *logLevelFlag).Info("Parsed flags, setting log level.")
	log.SetLevel(log.Level(*logLevelFlag))
	log.Info("Set logging level.")

	// Converting compression method flag:
	compressionMethod := uint16(*compressionMethodFlag)

	// Getting absolute path to input directory:
	absolutePathInputDirectory, _ := filepath.Abs(*inputDirectory)
	// absolutePathInterDirectory, _ := filepath.Abs(*interDirectory)
	absolutePathOutputDirectory, _ := filepath.Abs(*outputDirectory)

	log.WithFields(log.Fields{
		"inputDirectory":    absolutePathInputDirectory,
		"outputDirectory":   absolutePathOutputDirectory,
		"filesInPackage":    *filesInPackage,
		"compressionMethod": compressionMethod}).Info("Parsed command line flags")

	// Getting list of absolute paths for files from input directory:
	listOfInputFiles := listFiles(absolutePathInputDirectory, ".SC2Replay")

	// Register a custom compressor.
	zip.RegisterCompressor(12, func(out io.Writer) (io.WriteCloser, error) {
		return bzip2.NewWriterLevel(out, 9)
	})

	myProgressBar := progressbar.Default(int64(len(listOfInputFiles)))

	readErrorCounter := 0
	compressionErrorCounter := 0
	processedCounter := 0
	packageCounter := 0

	// Helper method returning bytes buffer and zip writer:
	buffer, writer := initBufferWriter()
	log.Info("Initialized buffer and writer.")

	for _, replayFile := range listOfInputFiles {

		didWork, replayString, replaySummary := dataproc.Pipeline(replayFile)
		if !didWork {
			readErrorCounter++
			continue
		}
		fmt.Println(replaySummary)

		// TODO: Handle replaySummary that is being created.
		// TODO: Write summary to JSON

		// Helper saving to zip archive:
		savedSuccess := saveFileToArchive(replayString, replayFile, compressionMethod, writer)
		if !savedSuccess {
			compressionErrorCounter++
			continue
		}
		log.Info("Added file to zip archive.")

		processedCounter++
		filesLeftToProcess := len(listOfInputFiles) - processedCounter
		// Remembering how much files were processed and created as .json:
		myProgressBar.Add(1)
		// Stop after reaching the limit and compress into a bzip2
		if processedCounter%*filesInPackage == 0 || filesLeftToProcess == 0 {
			log.Info("Detected processed counter to be within filesInPackage threshold.")
			writer.Close()
			packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(packageCounter)+".zip")
			_ = ioutil.WriteFile(packageAbsPath, buffer.Bytes(), 0777)
			log.Info("Saved package: %s to path: %s", packageCounter, packageAbsPath)
			packageCounter++

			// Helper method returning bytes buffer and zip writer:
			buffer, writer = initBufferWriter()
			log.Info("Initialized buffer and writer.")
		}
	}
	if readErrorCounter > 0 {
		log.WithField("readErrors", readErrorCounter).Info("Finished processing ", readErrorCounter)
	}
	if compressionErrorCounter > 0 {
		log.WithField("compressionErrors", compressionErrorCounter).Info("Finished processing compressionErrors: ", compressionErrorCounter)
	}
}
