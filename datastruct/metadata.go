package datastruct

// CleanedMetadata is a structure holding cleaned replay metadata derived from s2prot.Rep.Metadata
type CleanedMetadata struct {
	BaseBuild string `json:"baseBuild"`
	DataBuild string `json:"dataBuild"`
	// Duration    float64 `json:"durationSeconds"`
	GameVersion string `json:"gameVersion"`
	// Players     []CleanedPlayer `json:"players"`
	MapName string `json:"mapName"` // Originally Title
}

// CleanedPlayer is cleaned player information derived from s2prot.Rep
type CleanedPlayer struct {
	PlayerID     uint8  `json:"playerID"`
	APM          uint16 `json:"APM"`
	MMR          uint16 `json:"MMR"`
	Result       string `json:"result"`
	AssignedRace string `json:"assignedRace"`
	SelectedRace string `json:"selectedRace"`
}
