package utils

import (
	"os"
	"testing"
)

var TEST_FILE_PATH = "../test_files/"
var TEST_REPLAYS_PATH = "../test_files/test_replays"
var TEST_PROFILER_PATH = "../test_files/test_logs/test_profiler.txt"

func TestSetProfilingEmpty(t *testing.T) {

	_, profilingSetOk := SetProfiling("")

	if profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned true on an empty string!.")
	}
}

func TestSetProfiling(t *testing.T) {

	profilerFile, profilingSetOk := SetProfiling(TEST_PROFILER_PATH)

	if !profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned false on a valid path.")
	}

	err := profilerFile.Close()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't close the profiling file.")
	}

	err = os.Remove(TEST_PROFILER_PATH)
	if err != nil {
		t.Fatalf("Test Failed! Cannot delete profiling file.")
	}

}

func TestGetChunksOfFiles(t *testing.T) {

	// Read all the test input directory:
	sliceOfFiles := ListFiles(TEST_REPLAYS_PATH, ".SC2Replay")
	sliceOfChunks, getOk := GetChunksOfFiles(sliceOfFiles, 1)

	if !getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = false.")
	}

	if len(sliceOfChunks) != len(sliceOfFiles) {
		t.Fatalf("Test Failed! lenghts of slices mismatch.")
	}
}

func TestGetChunksOfFilesZero(t *testing.T) {

	// Read all the test input directory:
	sliceOfFiles := ListFiles(TEST_REPLAYS_PATH, ".SC2Replay")
	sliceOfChunks, getOk := GetChunksOfFiles(sliceOfFiles, 0)

	if !getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = false.")
	}

	if len(sliceOfChunks) != 1 {
		t.Fatalf("Test Failed! lenghts of slices mismatch.")
	}

}

func TestGetChunksOfFilesMinus(t *testing.T) {

	// Read all the test input directory:
	sliceOfFiles := ListFiles(TEST_REPLAYS_PATH, ".SC2Replay")
	_, getOk := GetChunksOfFiles(sliceOfFiles, -1)

	if getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = true.")
	}

}
