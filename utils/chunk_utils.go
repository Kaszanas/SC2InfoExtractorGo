package utils

import (
	"math"

	log "github.com/sirupsen/logrus"
)

// getChunksOfFiles returns chunks of files for processing.
func GetChunksOfFiles(slice []string, chunkSize int) ([][]string, bool) {

	log.Info("Entered chunkSlice()")

	if chunkSize < 0 {
		return [][]string{}, false
	}

	if chunkSize == 0 {
		return [][]string{slice}, true
	}

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
	return chunks, true
}

func GetChunkListAndPackageBool(
	listOfInputFiles []string,
	numberOfPackages int, numberOfThreads int,
	lenListOfInputFiles int) ([][]string, bool) {

	log.Info("Entered getChunkListAndPackageBool()")

	packageToZipBool := true
	if numberOfPackages == 0 {
		packageToZipBool = false
	}

	var numberOfFilesInPackage int
	if packageToZipBool {
		// If we package all of the replays into ZIP we use user specified number of packages. Number of chunks is n_files/n_user_provided_packages
		numberOfFilesInPackage = int(math.Ceil(float64(lenListOfInputFiles) / float64(numberOfPackages)))
		listOfChunksFiles, _ := GetChunksOfFiles(listOfInputFiles, numberOfFilesInPackage)
		return listOfChunksFiles, packageToZipBool
	}

	// If we write stringified .json files of replays to drive without packaging the number of chunks will be n_files/n_threads
	numberOfFilesInPackage = int(math.Ceil(float64(lenListOfInputFiles) / float64(numberOfThreads)))
	listOfChunksFiles, _ := GetChunksOfFiles(listOfInputFiles, numberOfFilesInPackage)

	return listOfChunksFiles, packageToZipBool

}
