package file_utils

import (
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/chunk_utils"
)

// TestGetChunksOfFiles tests the GetChunksOfFiles function.
func TestGetChunksOfFiles(t *testing.T) {
	testReplaysPath, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test input directory.")
	}

	// Read all the test input directory:
	sliceOfFiles, err := ListFiles(testReplaysPath, ".SC2Replay")
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the list of files.")
	}
	sliceOfChunks, getOk := chunk_utils.GetChunks(sliceOfFiles, 1)
	if !getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = false.")
	}

	if len(sliceOfChunks) != len(sliceOfFiles) {
		t.Fatalf("Test Failed! lenghts of slices mismatch.")
	}
}
