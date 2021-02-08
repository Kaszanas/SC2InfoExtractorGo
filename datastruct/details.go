package datastruct

import "time"

type CleanedDetails struct {
	GameSpeed       uint8
	IsBlizzardMap   bool
	PlayerList      []CleanedPlayerListStruct
	TimeLocalOffset time.Duration
	TimeUTC         time.Time
	MapName         string // originally title
}

type CleanedPlayerListStruct struct {
	Color    PlayerListColor
	Handicap uint8
	Name     string
	Race     string
	Result   uint8
	TeamID   uint8
	Realm    uint8
	Region   uint8
}

type PlayerListColor struct {
	A uint8
	B uint8
	G uint8
	R uint8
}
