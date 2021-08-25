package datastruct

import (
	"github.com/icza/s2prot"
)

// CleanedReplay is a structure holding cleaned data derived from s2prot.Rep
type CleanedReplay struct {
	Header            CleanedHeader                  `json:"header"`
	InitData          CleanedInitData                `json:"initData"`
	Details           CleanedDetails                 `json:"details"`
	Metadata          CleanedMetadata                `json:"metadata"`
	MessageEvents     []s2prot.Struct                `json:"messageEvents"`
	GameEvents        []s2prot.Struct                `json:"gameEvents"`
	TrackerEvents     []s2prot.Struct                `json:"trackerEvents"`
	ToonPlayerDescMap map[string]EnhancedToonDescMap `json:"ToonPlayerDescMap"` //map[string]*rep.PlayerDesc
	GameEvtsErr       bool                           `json:"gameEventsErr"`
	MessageEvtsErr    bool                           `json:"messageEventsErr"`
	TrackerEvtsErr    bool                           `json:"trackerEvtsErr"`
}

// EnhancedToonDescMap is a structure that provides more information that standard map[string]*rep.PlayerDesc
type EnhancedToonDescMap struct {
	Name                string          `json:"nickname"`
	PlayerID            int64           `json:"playerID"`
	UserID              int64           `json:"userID"`
	SQ                  int32           `json:"SQ"`
	SupplyCappedPercent int32           `json:"supplyCappedPercent"`
	StartDir            int32           `json:"startDir"`
	StartLocX           int64           `json:"startLocX"`
	StartLocY           int64           `json:"startLocY"`
	AssignedRace        string          `json:"race"`
	SelectedRace        string          `json:"selectedRace"`
	APM                 float64         `json:"APM"`
	MMR                 float64         `json:"MMR"`
	Result              string          `json:"result"`
	Region              string          `json:"region"`
	Realm               string          `json:"realm"`
	HighestLeague       string          `json:"highestLeague"`
	IsInClan            bool            `json:"isInClan"`
	ClanTag             string          `json:"clanTag"`
	Handicap            int64           `json:"handicap"`
	Color               PlayerListColor `json:"color"`
}
