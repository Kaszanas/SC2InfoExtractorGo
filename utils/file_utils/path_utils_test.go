package file_utils

import (
	"os"
	"path/filepath"
	"slices"
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

// TestListFilesExtensionCaseInsensitive tests that ListFiles matches
// the filter extension regardless of letter case.
func TestListFilesExtensionCaseInsensitive(t *testing.T) {
	inputDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(inputDir, "sub"), 0755); err != nil {
		t.Fatalf("Test Failed! Couldn't create the subdirectory: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(inputDir, "a.SC2Replay"),
		filepath.Join(inputDir, "b.sc2replay"),
		filepath.Join(inputDir, "c.SC2REPLAY"),
		filepath.Join(inputDir, "sub", "d.Sc2rEpLaY"),
	}
	otherFiles := []string{
		filepath.Join(inputDir, "e.json"),
		filepath.Join(inputDir, "f.SC2Replay.bak"),
		filepath.Join(inputDir, "g"),
	}
	for _, file := range append(slices.Clone(expectedFiles), otherFiles...) {
		if err := os.WriteFile(file, nil, 0644); err != nil {
			t.Fatalf("Test Failed! Couldn't create %s: %v", file, err)
		}
	}

	sliceOfFiles, err := ListFiles(inputDir, ".SC2Replay")
	if err != nil {
		t.Fatalf("Test Failed! Couldn't get the list of files: %v", err)
	}

	slices.Sort(sliceOfFiles)
	slices.Sort(expectedFiles)
	if !slices.Equal(sliceOfFiles, expectedFiles) {
		t.Fatalf("Test Failed! got %v, want %v", sliceOfFiles, expectedFiles)
	}
}
