package main

import (
	"fmt"
	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"archive/zip"
	"flag"
	"github.com/larzconwell/bzip2"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

func main() {

	// Command line arguments:
	inputDirectory := flag.String("input", "./DEMOS/Input", "Input directory where .SC2Replay files are held.")
	// interDirectory := flag.String("inter", "./Demos/Intermediate", "Intermediate directory where .json files will be stored before bzip2 compression.")
	outputDirectory := flag.String("output", "./DEMOS/Output", "Output directory where compressed bzip2 packages will be stored.")
	filesInPackage := flag.Int("files_in_package", 3, "Provide a number of files to be compressed into a bzip2 archive.")
	// Other compression methods than Deflate need to be registered further down in the code:
	compressionMethodFlag := flag.Int("compression_method", 8, "Provide a compression method number, default is 8 'Deflate', other compression methods need to be registered in code.")

	flag.Parse()

	// Converting compression method flag:
	compressionMethod := uint16(*compressionMethodFlag)

	// Getting absolute path to input directory:
	absolutePathInputDirectory, _ := filepath.Abs(*inputDirectory)
	// absolutePathInterDirectory, _ := filepath.Abs(*interDirectory)
	absolutePathOutputDirectory, _ := filepath.Abs(*outputDirectory)
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

	for _, replayFile := range listOfInputFiles {

		didWork, replayString := stringifyReplay(replayFile)
		if !didWork {
			readErrorCounter++
			continue
		}

		// TODO: Write summary to CSV or JSON or sqlite

		// Helper saving to zip archive:
		saveFileToArchive(replayString, replayFile, compressionMethod, writer)

		processedCounter++
		filesLeftToProcess := len(listOfInputFiles) - processedCounter
		// Remembering how much files were processed and created as .json:
		myProgressBar.Add(1)
		// Stop after reaching the limit and compress into a bzip2
		if processedCounter%*filesInPackage == 0 || filesLeftToProcess == 0 {
			writer.Close()
			packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(packageCounter)+".zip")
			_ = ioutil.WriteFile(packageAbsPath, buffer.Bytes(), 0777)
			packageCounter++

			// Helper method returning bytes buffer and zip writer:
			buffer, writer = initBufferWriter()
		}

	}
	fmt.Println(readErrorCounter)
	fmt.Println(compressionErrorCounter)
}
