package datastruct

import "time"

// CleanedDetails is information about SC2 replay details derived from s2prot.Rep
type CleanedDetails struct {
	GameSpeed       uint8                     `json:"gameSpeed"`
	IsBlizzardMap   bool                      `json:"isBlizzardMap"`
	PlayerList      []CleanedPlayerListStruct `json:"playerList"`
	TimeLocalOffset time.Duration             `json:"timeLocalOffset"`
	TimeUTC         time.Time                 `json:"timeUTC"`
	MapName         string                    `json:"mapName"` // originally title
}

// CleanedPlayerListStruct is a nested structure that lies within CleanedDetails derived from s2prot.Rep
type CleanedPlayerListStruct struct {
	Color    PlayerListColor `json:"color"`
	Handicap uint8           `json:"handicap"`
	Name     string          `json:"name"`
	Race     string          `json:"race"`
	Result   uint8           `json:"result"`
	TeamID   uint8           `json:"teamID"`
	Realm    uint8           `json:"realm"`
	Region   uint8           `json:"region"`
}

// PlayerListColor is a color information of the player derived from s2prot.Rep
type PlayerListColor struct {
	A uint8 `json:"a"`
	B uint8 `json:"b"`
	G uint8 `json:"g"`
	R uint8 `json:"r"`
}
