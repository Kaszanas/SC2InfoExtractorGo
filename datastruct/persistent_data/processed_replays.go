package persistent_data

import "github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"

// ProcessedReplaysToMaps
type ProcessedReplaysToMaps struct {
	ProcessedFiles map[string]interface{} `json:"processedReplays"`
}

// NewProcessedReplaysToMaps returns empty ProcessingIngo struct.
func NewProcessedReplaysToMaps(filepath string) ProcessedReplaysToMaps {

	// check if the file exists:
	mapToPopulate := make(map[string]interface{})
	file_utils.ReadOrCreateFile("processed_replays.json")
	file_utils.UnmarshalJsonFile(filepath, &mapToPopulate)

	return ProcessedReplaysToMaps{
		ProcessedFiles: mapToPopulate,
	}
}
