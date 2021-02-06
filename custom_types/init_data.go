package main

import (
	"github.com/icza/s2prot"
)

type CleanedInitData struct {
	gameDescription CleanedGameDescription
	userInitData    []CleanedUserInitData
}

type CleanedGameDescription struct {
	gameOptions         s2prot.Struct
	gameSpeed           uint8
	isBlizzardMap       bool
	mapAuthorName       string
	mapFileSyncChecksum string
	mapSizeX            uint32
	mapSizeY            uint32
	maxPlayers          uint8
}

type CleanedUserInitData struct {
	combinedRaceLevels uint64
	highestLeague      uint32
	name               string
	isInClan           bool
}
