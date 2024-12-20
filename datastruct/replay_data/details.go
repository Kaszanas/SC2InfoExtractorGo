package replay_data

import "time"

// CleanedDetails is a structure holding information
// about SC2 replay details derived from s2prot.Rep.Details
type CleanedDetails struct {
	GameSpeed     string `json:"gameSpeed"`
	IsBlizzardMap bool   `json:"isBlizzardMap"`
	// PlayerList    []CleanedPlayerListStruct `json:"playerList"`
	// TimeLocalOffset time.Duration             `json:"timeLocalOffset"`
	TimeUTC time.Time `json:"timeUTC"`
	// MapName string    `json:"mapName"` // originally title
}

// CleanedPlayerListStruct is a nested structure that
// lies within CleanedDetails derived from s2prot.Rep.Details.PlayerList
type CleanedPlayerListStruct struct {
	Name               string          `json:"name"`
	Race               rune            `json:"race"`
	Result             string          `json:"result"`
	IsInClan           bool            `json:"isInClan"`
	HighestLeague      string          `json:"highestLeague"`
	Handicap           uint8           `json:"handicap"`
	TeamID             int64           `json:"teamID"`
	Region             string          `json:"region"`
	Realm              string          `json:"realm"`
	CombinedRaceLevels uint64          `json:"combinedRaceLevels"`
	Color              PlayerListColor `json:"color"`
}

// PlayerListColor is a color information of the player derived from s2prot.Rep
type PlayerListColor struct {
	A uint8 `json:"a"`
	B uint8 `json:"b"`
	G uint8 `json:"g"`
	R uint8 `json:"r"`
}
