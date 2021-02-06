package main

import "time"

type CleanedDetails struct {
	gameSpeed       uint8
	isBlizzardMap   bool
	playerList      []CleanedPlayerListStruct
	timeLocalOffset time.Duration
	timeUTC         time.Time
	mapName         string // originally title
}

type CleanedPlayerListStruct struct {
	color    PlayerListColor
	handicap uint8
	name     string
	race     string
	result   uint8
	teamID   uint8
	realm    uint8
	region   uint8
}

type PlayerListColor struct {
	a uint16
	b uint16
	g uint16
	r uint16
}
