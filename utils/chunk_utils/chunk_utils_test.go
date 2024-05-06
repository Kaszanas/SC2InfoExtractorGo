package chunk_utils

import (
	"testing"
)

// TestGetChunksOfFiles tests the GetChunksOfFiles function by passing a zero number of chunks.
// This should return a single chunk with all the files.
func TestGetChunksOfFilesZero(t *testing.T) {

	// Read all the test input directory:
	sliceOfFiles := []string{"test_file.txt"}
	sliceOfChunks, getOk := GetChunksOfFiles(sliceOfFiles, 0)

	if !getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = false.")
	}

	if len(sliceOfChunks) != 1 {
		t.Fatalf("Test Failed! lenghts of slices mismatch.")
	}

}

// TestGetChunksOfFilesMinus tests the GetChunksOfFiles function
// by passing a negative number of chunks.
func TestGetChunksOfFilesMinus(t *testing.T) {

	// Read all the test input directory:
	sliceOfFiles := []string{"test_file.txt"}
	_, getOk := GetChunksOfFiles(sliceOfFiles, -1)

	if getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = true.")
	}

}
