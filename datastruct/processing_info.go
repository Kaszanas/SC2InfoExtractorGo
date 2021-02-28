package datastruct

// ProcessingInfoMap holds information that is used to create processing.log, which is anonymizedPlayers in a persistent map from toon to unique integer, slice of processed files so that there is a state of all of the processed files.
type ProcessingInfo struct {
	AnonymizedPlayers map[string]int `json:"anonymizedPlayers"`
	ProcessedFiles    []string       `json:"processedFiles"`
}
