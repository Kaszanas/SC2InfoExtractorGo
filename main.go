package main

import (
	"fmt"
	// "github.com/icza/mpq"
	"encoding/json"
	// "github.com/icza/s2prot"
	"archive/zip"
	"bytes"
	"flag"
	"github.com/icza/s2prot/rep"
	"github.com/larzconwell/bzip2"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
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
	var listToCompress []string
	for _, replayFile := range listOfInputFiles {

		replayData, err := rep.NewFromFile(replayFile)

		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			readErrorCounter++
			continue
		}
		defer replayData.Close()

		header := replayData.Header.String()
		details := replayData.Details.String()
		initData := replayData.InitData.String()
		attrEvts := replayData.AttrEvts.String()
		metadata := replayData.Metadata.String()

		PIDPlayerDescMap := replayData.TrackerEvts.PIDPlayerDescMap
		ToonPlayerDescMap := replayData.TrackerEvts.ToonPlayerDescMap

		// Creating lists of strings for later use in generating JSON out of the replay data:
		var gameEventStrings []string
		for _, gameEvent := range replayData.GameEvts {
			gameEventStrings = append(gameEventStrings, gameEvent.String())
		}

		var messageEventStrings []string
		for _, messageEvent := range replayData.MessageEvts {
			messageEventStrings = append(messageEventStrings, messageEvent.String())
		}

		var trackerEventStrings []string
		for _, trackerEvent := range replayData.TrackerEvts.Evts {
			trackerEventStrings = append(trackerEventStrings, trackerEvent.String())
		}

		// These structures are handled differently as it is a Map without .String() method:
		var PIDPlayerDescMapStrings []string
		for PIDPlayerDescKey, PIDPlayerDescValue := range PIDPlayerDescMap {

			// Converting ID to string:
			playerNumber := strconv.FormatInt(PIDPlayerDescKey, 10)

			// Converting struct to JSON:
			playerDescInformation, err := json.Marshal(PIDPlayerDescValue)

			if err != nil {
				panic(err)
			}

			// Putting everything together:
			PIDPlayerDescMapStrings = append(PIDPlayerDescMapStrings, "\""+playerNumber+"\": "+string(playerDescInformation))
		}

		var ToonPlayerDescMapStrings []string
		for ToonPlayerDescKey, ToonPlayerDescValue := range ToonPlayerDescMap {

			// Converting ID to string:
			playerToon := ToonPlayerDescKey

			// Converting struct to JSON:
			playerDescInformation, err := json.Marshal(ToonPlayerDescValue)

			if err != nil {
				panic(err)
			}

			// Putting everything together:
			ToonPlayerDescMapStrings = append(ToonPlayerDescMapStrings, "\""+playerToon+"\": "+string(playerDescInformation))
		}

		// Booleans saying if processing had any errors
		gameEvtsErr := strconv.FormatBool(replayData.GameEvtsErr)
		messageEvtsErr := strconv.FormatBool(replayData.MessageEvtsErr)
		trackerEvtsErr := strconv.FormatBool(replayData.TrackerEvtsErr)

		// Crezting JSON structure by hand:
		var strBuilder strings.Builder
		fmt.Fprintf(&strBuilder, "{\n")
		fmt.Fprintf(&strBuilder, "  \"header\": %s,\n", header)
		fmt.Fprintf(&strBuilder, "  \"initData\": %s,\n", initData)
		fmt.Fprintf(&strBuilder, "  \"details\": %s,\n", details)
		fmt.Fprintf(&strBuilder, "  \"attrEvts\": %s,\n", attrEvts)
		fmt.Fprintf(&strBuilder, "  \"metadata\": %s,\n", metadata)
		fmt.Fprintf(&strBuilder, "  \"gameEvtsErr\": %s,\n", gameEvtsErr)
		fmt.Fprintf(&strBuilder, "  \"messageEvtsErr\": %s,\n", messageEvtsErr)
		fmt.Fprintf(&strBuilder, "  \"trackerEvtsErr\": %s,\n", trackerEvtsErr)
		fmt.Fprintf(&strBuilder, "  \"messageEventsStrings\": [%s],\n", strings.Join(messageEventStrings, ",\n"))
		fmt.Fprintf(&strBuilder, "  \"gameEventStrings\": [%s],\n", strings.Join(gameEventStrings, ",\n"))
		fmt.Fprintf(&strBuilder, "  \"trackerEventStrings\": [%s],\n", strings.Join(trackerEventStrings, ",\n"))
		fmt.Fprintf(&strBuilder, "  \"PIDPlayerDescMap\": {%s},\n", strings.Join(PIDPlayerDescMapStrings, ",\n"))
		fmt.Fprintf(&strBuilder, "  \"ToonPlayerDescMap\": {%s},\n", strings.Join(ToonPlayerDescMapStrings, ",\n"))
		fmt.Fprintf(&strBuilder, "  \"gameEvtsErr\": %s", gameEvtsErr+",\n")
		fmt.Fprintf(&strBuilder, "  \"messageEvtsErr\": %s", messageEvtsErr+",\n")
		fmt.Fprintf(&strBuilder, "  \"trackerEvtsErr\": %s", trackerEvtsErr+"\n")
		fmt.Fprintf(&strBuilder, "  ")
		fmt.Fprintf(&strBuilder, "}")

		// _, replayFilename := filepath.Split(replayFile)
		// finalFilename := strings.TrimSuffix(replayFilename, filepath.Ext(replayFilename)) + ".json"

		listToCompress = append(listToCompress, strBuilder.String())
		// Writing JSON file:
		// _ = ioutil.WriteFile(filepath.Join(absolutePathInterDirectory, finalFilename), []byte(strBuilder.String()), 0644)
		// Remembering how much files were processed and created as .json:
		myProgressBar.Add(1)

		filesLeftToProcess := len(listOfInputFiles) - processedCounter

		// Create a buffer to write our archive to:
		buf := new(bytes.Buffer)

		// Create a new zip archive:
		w := zip.NewWriter(buf)

		// Read byte array from json file:
		// JSONContents, err := ioutil.ReadFile(file)
		// if err != nil {
		// 	fmt.Printf("Failed to open %s: %s", file, err)
		// 	panic("Error")
		// }

		// TODO: This might be needed if string will be read from memory:
		// Converting from string to bytes:
		// someStringBytes := []byte(someString)

		jsonBytes := []byte(strBuilder.String())
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
		w.Close()
		processedCounter++

		// Stop after reaching the limit and compress into a bzip2
		if processedCounter%*filesInPackage == 0 || filesLeftToProcess == 0 {

			// Create empty zip file with numbered filename.
			// emptyZip, err := os.Create(filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(packageCounter)+".zip"))
			// if err != nil {
			// 	panic(err)
			// }

			// _, compressionError := bzipWriter.Write(JSONContents)
			// if compressionError != nil {
			// 	fmt.Printf("Failed to write %s to zip: %s", file, err)
			// 	compressionErrorCounter++
			// }

			// Delete intermediate .json files
			// dir, err := ioutil.ReadDir(absolutePathInterDirectory)
			// for _, d := range dir {
			// 	os.RemoveAll(filepath.Join([]string{"tmp", d.Name()}...))
			// }

			packageAbsPath := filepath.Join(absolutePathOutputDirectory, "package_"+strconv.Itoa(packageCounter)+".zip")
			_ = ioutil.WriteFile(packageAbsPath, buf.Bytes(), 0777)
			packageCounter++
		}

	}
	fmt.Println(readErrorCounter)
	fmt.Println(compressionErrorCounter)
}
