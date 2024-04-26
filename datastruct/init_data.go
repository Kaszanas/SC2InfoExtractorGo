package datastruct

import (
	"github.com/icza/s2prot"
)

// CleanedInitData is a structure holding cleaned
// initial data of a replay derived from s2prot.Rep.initData
type CleanedInitData struct {
	GameDescription CleanedGameDescription `json:"gameDescription"`
}

// CleanedGameDescription is cleaned game description
// of a replay derived from s2prot.Rep.initData.GameDescription
type CleanedGameDescription struct {
	GameOptions         s2prot.Struct `json:"gameOptions"`
	GameSpeed           string        `json:"gameSpeed"`
	IsBlizzardMap       bool          `json:"isBlizzardMap"`
	MapAuthorName       string        `json:"mapAuthorName"`
	MapFileSyncChecksum int64         `json:"mapFileSyncChecksum"`
	MapSizeX            uint32        `json:"mapSizeX"`
	MapSizeY            uint32        `json:"mapSizeY"`
	MaxPlayers          uint8         `json:"maxPlayers"`
}

// CleanedUserInitData is cleaned user initial data of a replay derived from s2prot.Rep
type CleanedUserInitData struct {
	CombinedRaceLevels uint64 `json:"combinedRaceLevels"`
	HighestLeague      string `json:"highestLeague"`
	Name               string `json:"name"`
	ClanTag            string `json:"clanTag"`
	IsInClan           bool   `json:"isInClan"`
}
