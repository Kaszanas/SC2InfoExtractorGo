package datastruct

// ProcessingInfo is a structure holding information that
// is used to create processing.log, which is anonymizedPlayers
// in a persistent map from toon to unique integer,
// slice of processed files so that there is a state of all of the processed files.
type ProcessingInfo struct {
	ProcessedFiles  []string            `json:"processedFiles"`
	FailedToProcess []map[string]string `json:"failedToProcess"`
}

// NewProcessingInfo returns empty ProcessingIngo struct.
func NewProcessingInfo() ProcessingInfo {
	return ProcessingInfo{
		ProcessedFiles:  make([]string, 0),
		FailedToProcess: make([]map[string]string, 0),
	}
}
