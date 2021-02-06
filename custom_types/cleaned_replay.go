package main

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

type CleanedReplay struct {
	header            CleanedHeader
	initData          CleanedInitData
	details           CleanedDetails
	metadata          CleanedMetadata
	messageEvents     []s2prot.Event
	gameEvents        []s2prot.Event
	trackerEvents     []s2prot.Event
	PIDPlayerDescMap  map[string]*rep.PlayerDesc
	toonPlayerDescMap map[string]*rep.PlayerDesc
	gameEvtsErr       bool
	messageEvtsErr    bool
	trackerEvtsErr    bool
}
