package datastruct

import "time"

type CleanedMetadata struct {
	BaseBuild   string
	DataBuild   string
	Duration    time.Duration
	GameVersion string
	Players     []CleanedPlayers
	MapName     string // Originally Title
}

type CleanedPlayers struct {
	PlayerID     uint8
	APM          uint16
	MMR          uint16
	Result       string
	AssignedRace string
	SelectedRace string
}
