package utils

import (
	"os"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
)

func TestSetProfilingEmpty(t *testing.T) {

	_, profilingSetOk := SetProfiling("")

	if profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned true on an empty string!.")
	}
}

func TestSetProfiling(t *testing.T) {

	testProfilerPath, err := settings.GetProfilerReportPath()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test profiler path.")
	}

	profilerFile, profilingSetOk := SetProfiling(testProfilerPath)

	if !profilingSetOk {
		t.Fatalf("Test Failed! setProfiling returned false on a valid path.")
	}

	err = profilerFile.Close()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't close the profiling file.")
	}

	err = os.Remove(testProfilerPath)
	if err != nil {
		t.Fatalf("Test Failed! Cannot delete profiling file.")
	}

}

func TestGetChunksOfFiles(t *testing.T) {
	testReplaysPath, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test input directory.")
	}

	// Read all the test input directory:
	sliceOfFiles := ListFiles(testReplaysPath, ".SC2Replay")
	sliceOfChunks, getOk := GetChunksOfFiles(sliceOfFiles, 1)

	if !getOk {
		t.Fatalf("Test Failed! getChunksOfFiles() returned getOk = false.")
	}

	if len(sliceOfChunks) != len(sliceOfFiles) {
		t.Fatalf("Test Failed! lenghts of slices mismatch.")
	}
}

func TestGetChunksOfFilesZero(t *testing.T) {

	testReplaysPath, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the test input directory.")
	}

	// Read all the test input directory:
	sliceOfFiles := ListFiles(testReplaysPath, ".SC2Replay")
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
