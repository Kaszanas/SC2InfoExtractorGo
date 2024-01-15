package utils

import (
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

func TestGetChunksOfFilesZero(t *testing.T) {

	testReplaysPath, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test input directory.")
	}

	// Read all the test input directory:
	sliceOfFiles := ListFiles(testReplaysPath, ".SC2Replay")
	// TODO: Split this from getting SC2Replay files just pass a list of strings representing files.
	sliceOfChunks, getOk := GetChunksOfFiles(sliceOfFiles, 0)

	if !getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = false.")
	}

	if len(sliceOfChunks) != 1 {
		t.Fatalf("Test Failed! lenghts of slices mismatch.")
	}

}

func TestGetChunksOfFilesMinus(t *testing.T) {
	testReplaysPath, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test input directory.")
	}

	// Read all the test input directory:
	sliceOfFiles := ListFiles(testReplaysPath, ".SC2Replay")
	_, getOk := GetChunksOfFiles(sliceOfFiles, -1)

	if getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = true.")
	}

}
