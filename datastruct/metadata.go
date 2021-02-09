package datastruct

import "time"

type CleanedMetadata struct {
	BaseBuild   string
	DataBuild   string
	Duration    time.Duration
	GameVersion string
	Players     []CleanedPlayer
	MapName     string // Originally Title
}

type CleanedPlayer struct {
	PlayerID     uint8
	APM          uint16
	MMR          uint16
	Result       string
	AssignedRace string
	SelectedRace string
}
