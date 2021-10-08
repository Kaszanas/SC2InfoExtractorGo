package dataproc

import (
	"testing"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

func TestPipelineWrapper(t *testing.T) {

	testOutputPath := t.TempDir()

	testLocalizationMapFile := "../test_files/test_map_mapping/output.json"
	logsFilepath := "../test_files/test_logs/"

	testReplayDir := "../test_files/test_replays"
	sliceOfFiles := utils.ListFiles(testReplayDir, ".SC2Replay")
	chunks, getOk := utils.GetChunksOfFiles(sliceOfFiles, 5)

	if !getOk {
		t.Fatalf("Test Failed! Could not produce chunks of files!")
	}

	packageToZip := false
	integrityCheck := true
	validityCheck := false
	performFilteringBool := false
	gameModeCheckFlag := 0
	performPlayerAnonymization := false
	performChatAnonymization := false
	performCleanup := true
	compressionMethod := uint16(8)
	numberOfThreads := 1

	localizedMapsMap := map[string]interface{}(nil)
	localizedMapsMap = utils.UnmarshalLocaleMapping(testLocalizationMapFile)
	if localizedMapsMap == nil {
		log.Fatalf("Test Failed! Could not unmarshall the localization mapping file.")
	}

	PipelineWrapper(
		testOutputPath,
		chunks,
		packageToZip,
		integrityCheck,
		validityCheck,
		performFilteringBool,
		gameModeCheckFlag,
		performPlayerAnonymization,
		performChatAnonymization,
		performCleanup,
		localizedMapsMap,
		compressionMethod,
		numberOfThreads,
		logsFilepath,
	)

}
