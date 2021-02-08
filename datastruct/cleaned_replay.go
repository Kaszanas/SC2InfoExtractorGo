package datastruct

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

type CleanedReplay struct {
	Header            CleanedHeader
	InitData          CleanedInitData
	Details           CleanedDetails
	Metadata          CleanedMetadata
	MessageEvents     []s2prot.Event
	GameEvents        []s2prot.Event
	TrackerEvents     []s2prot.Event
	PIDPlayerDescMap  map[int64]*rep.PlayerDesc
	ToonPlayerDescMap map[string]*rep.PlayerDesc
	GameEvtsErr       bool
	MessageEvtsErr    bool
	TrackerEvtsErr    bool
}
