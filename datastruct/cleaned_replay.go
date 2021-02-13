package datastruct

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
)

// CleanedReplay is a structure holding cleaned data derived from s2prot.Rep
type CleanedReplay struct {
	Header            CleanedHeader              `json:"header"`
	InitData          CleanedInitData            `json:"initData"`
	Details           CleanedDetails             `json:"details"`
	Metadata          CleanedMetadata            `json:"metadata"`
	MessageEvents     []s2prot.Event             `json:"messageEvents"`
	GameEvents        []s2prot.Event             `json:"gameEvents"`
	TrackerEvents     []s2prot.Event             `json:"trackerEvents"`
	PIDPlayerDescMap  map[int64]*rep.PlayerDesc  `json:"PIDPlayerDescMap"`
	ToonPlayerDescMap map[string]*rep.PlayerDesc `json:"ToonPlayerDescMap"`
	GameEvtsErr       bool                       `json:"gameEventsErr"`
	MessageEvtsErr    bool                       `json:"messageEventsErr"`
	TrackerEvtsErr    bool                       `json:"trackerEvtsErr"`
}
