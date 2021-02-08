package datastruct

import (
	"github.com/icza/s2prot"
)

type CleanedInitData struct {
	GameDescription CleanedGameDescription
	UserInitData    []CleanedUserInitData
}

type CleanedGameDescription struct {
	GameOptions         s2prot.Struct
	GameSpeed           uint8
	IsBlizzardMap       bool
	MapAuthorName       string
	MapFileSyncChecksum int
	MapSizeX            uint32
	MapSizeY            uint32
	MaxPlayers          uint8
}

type CleanedUserInitData struct {
	CombinedRaceLevels uint64
	HighestLeague      uint32
	Name               string
	IsInClan           bool
}
