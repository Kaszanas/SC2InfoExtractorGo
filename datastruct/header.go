package datastruct

import (
	"time"

	"github.com/icza/s2prot"
)

// CleanedHeader is a structure holding header information of a replay file derived from s2prot.Rep.Header
type CleanedHeader struct {
	ElapsedGameLoops uint64        `json:"elapsedGameLoops"`
	Duration         time.Duration `json:"duration"`
	UseScaledTime    bool          `json:"useScaledTime"`
	Version          s2prot.Struct `json:"version"`
}
