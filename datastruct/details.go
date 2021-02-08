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
	A uint16
	B uint16
	G uint16
	R uint16
}
