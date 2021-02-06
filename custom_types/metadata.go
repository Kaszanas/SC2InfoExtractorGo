package main

import "time"

type CleanedMetadata struct {
	baseBuild   string
	dataBuild   string
	duration    time.Duration
	gameVersion string
	players     []CleanedPlayers
	mapName     string // Originally Title
}

type CleanedPlayers struct {
	playerID     uint8
	APM          uint16
	MMR          uint16
	result       string
	assignedRace string
	selectedRace string
}
