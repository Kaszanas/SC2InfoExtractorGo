package dataproc

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/dataproc/downloader"
	"github.com/Kaszanas/SC2InfoExtractorGo/settings"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
)

// TestProcessReplayToJSON processes a single fixture replay and checks
// that the result is a JSON replay document.
func TestProcessReplayToJSON(t *testing.T) {
	testInputDir, err := settings.GetTestInputDirectory()
	if err != nil {
		t.Fatalf("Could not get the test input directory: %v", err)
	}
	replayFiles, err := file_utils.ListFiles(
		filepath.Join(testInputDir, "2016_IEM_10_Taipei"), ".SC2Replay")
	if err != nil || len(replayFiles) == 0 {
		t.Skip("Fixture replays not available, run `make fetch_test_fixtures`.")
	}
	replayFile := replayFiles[0]

	flags := utils.CLIFlags{
		SkipDependencyDownload: true,
		DependencyDirectory:    "../dependencies/",
		NumberOfThreads:        1,
		PerformIntegrityCheck:  true,
		PerformCleanup:         true,
		FilterGameMode:         0b11111111,
	}
	foreignToEnglishMapping := downloader.DependencyDownloaderPipeline(
		[]string{replayFile},
		filepath.Join(t.TempDir(), "map_foreign_to_english_mapping.json"),
		flags,
	)

	replayJSON, err := ProcessReplayToJSON(replayFile, foreignToEnglishMapping, flags)
	if err != nil {
		t.Fatalf("ProcessReplayToJSON returned error: %v", err)
	}

	var replay struct {
		AdditionalInformation struct {
			Filename string `json:"filename"`
		} `json:"additional_information"`
		Header map[string]any `json:"header"`
	}
	if err := json.Unmarshal([]byte(replayJSON), &replay); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if replay.AdditionalInformation.Filename != filepath.Base(replayFile) {
		t.Fatalf("additional_information.filename = %q, want %q",
			replay.AdditionalInformation.Filename, filepath.Base(replayFile))
	}
	if len(replay.Header) == 0 {
		t.Fatal("header is empty")
	}
}

// TestProcessReplayToJSONMissingFile checks that a replay which cannot be
// read is reported as an error.
func TestProcessReplayToJSONMissingFile(t *testing.T) {
	missingReplay := filepath.Join(t.TempDir(), "missing.SC2Replay")

	_, err := ProcessReplayToJSON(missingReplay, map[string]string{}, utils.CLIFlags{})
	if err == nil {
		t.Fatal("ProcessReplayToJSON succeeded for a missing file, want an error")
	}
}
