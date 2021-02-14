package datastruct

import (
	"github.com/icza/s2prot"
)

// CleanedInitData is cleaned initial data of a replay derived from s2prot.Rep
type CleanedInitData struct {
	GameDescription CleanedGameDescription `json:"gameDescription"`
	UserInitData    []CleanedUserInitData  `json:"userInitData"`
}

// CleanedGameDescription is cleaned game description of a replay derived from s2prot.Rep
type CleanedGameDescription struct {
	GameOptions         s2prot.Struct `json:"gameOptions"`
	GameSpeed           uint8         `json:"gameSpeed"`
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
	HighestLeague      uint32 `json:"highestLeague"`
	Name               string `json:"name"`
	IsInClan           bool   `json:"isInClan"`
}
