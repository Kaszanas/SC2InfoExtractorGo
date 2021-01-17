package main

import (
	"fmt"
	// "github.com/icza/mpq"
	// "github.com/icza/s2prot"
	"archive/zip"
	"bytes"
	"flag"
	"github.com/larzconwell/bzip2"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"
)

func main() {

	// Command line arguments:
	inputDirectory := flag.String("input", "./DEMOS/Input", "Input directory where .SC2Replay files are held.")
	// interDirectory := flag.String("inter", "./Demos/Intermediate", "Intermediate directory where .json files will be stored before bzip2 compression.")
	outputDirectory := flag.String("output", "./DEMOS/Output", "Output directory where compressed bzip2 packages will be stored.")
	filesInPackage := flag.Int("files_in_package", 3, "Provide a number of files to be compressed into a bzip2 archive.")
	// Other compression methods than Deflate need to be registered further down in the code:
	compressionMethod := flag.Int("compression_method", 8, "Provide a compression method number, default is 8 'Deflate', other compression methods need to be registered in code.")

	flag.Parse()

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

	// TODO: Helper function returning buffer and zip writer:
	// Create a buffer to write our archive to:
	buf := new(bytes.Buffer)
	// Create a new zip archive:
	w := zip.NewWriter(buf)

	for _, replayFile := range listOfInputFiles {

		didWork, replayString := stringifyReplay(replayFile)
		if !didWork {
			readErrorCounter++
			continue
		}

		// TODO: Write summary to CSV or JSON or sqlite

		// TODO: Helper that takes string, filename, zip file, and saves to zip archive under filename
		jsonBytes := []byte(replayString)
		_, fileHeaderFilename := filepath.Split(replayFile)

		fh := &zip.FileHeader{
			Name:               filepath.Base(fileHeaderFilename) + ".json",
			UncompressedSize64: uint64(len(jsonBytes)),
			Method:             uint16(*compressionMethod),
			Modified:           time.Now(),
		}
		fh.SetMode(0777)
		fw, err := w.CreateHeader(fh)

		if err != nil {
			fmt.Printf("Error: %s", err)
			panic("Error")
		}

		fw.Write(jsonBytes)
		// Up to here

		processedCounter++
		filesLeftToProcess := len(listOfInputFiles) - processedCounter
		// Remembering how much files were processed and created as .json:
		myProgressBar.Add(1)
		// Stop after reaching the limit and compress into a bzip2
		if processedCounter%*filesInPackage == 0 || filesLeftToProcess == 0 {
			//  TODO: Helper taking buffer and writing to drive
			w.Close()
			packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(packageCounter)+".zip")
			_ = ioutil.WriteFile(packageAbsPath, buf.Bytes(), 0777)
			packageCounter++

			// TODO: use previous helper:
			// Create a buffer to write our archive to:
			buf = new(bytes.Buffer)

			// Create a new zip archive:
			w = zip.NewWriter(buf)
		}

	}
	fmt.Println(readErrorCounter)
	fmt.Println(compressionErrorCounter)
}
