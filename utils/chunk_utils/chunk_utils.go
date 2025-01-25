package chunk_utils

import (
	"math"

	log "github.com/sirupsen/logrus"
)

// GetChunksOfFiles returns chunks of files for processing.
// GetChunks returns chunks of any type for processing.
func GetChunks[T any](slice []T, chunkSize int) ([][]T, bool) {
	log.Info("Entered GetChunks()")

	if chunkSize < 0 {
		return [][]T{}, false
	}

	if chunkSize == 0 {
		return [][]T{slice}, true
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond slice capacity:
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	log.Info("Finished GetChunks(), returning")
	return chunks, true
}

// GetChunkListAndPackageBool returns list of chunks of files that
// will be processed and a boolean specifying if the chunking process was a success.
func GetChunkListAndPackageBool[T any](
	listOfInputs []T,
	numberOfPackages int,
	numberOfThreads int,
	lenListOfInputFiles int,
) ([][]T, bool) {

	log.Info("Entered getChunkListAndPackageBool()")

	packageToZipBool := true
	if numberOfPackages == 0 {
		packageToZipBool = false
	}

	var numberOfFilesInPackage int
	if packageToZipBool {
		// If we package all of the replays into ZIP we use user
		// specified number of packages.
		// Number of chunks is n_files/n_user_provided_packages
		numberOfFilesInPackage = int(math.Ceil(float64(lenListOfInputFiles) / float64(numberOfPackages)))
		listOfChunksFiles, _ := GetChunks(listOfInputs, numberOfFilesInPackage)
		return listOfChunksFiles, packageToZipBool
	}

	// If we write stringified .json files of replays to drive without
	// packaging the number of chunks will be n_files/n_threads
	numberOfFilesInPackage = int(math.Ceil(float64(lenListOfInputFiles) / float64(numberOfThreads)))
	listOfChunksFiles, _ := GetChunks(listOfInputs, numberOfFilesInPackage)

	return listOfChunksFiles, packageToZipBool
}
