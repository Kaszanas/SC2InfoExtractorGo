package datastruct

import (
	"github.com/icza/s2prot"
)

// CleanedHeader is a structure holding header information of a replay file derived from s2prot.Rep.Header
type CleanedHeader struct {
	ElapsedGameLoops uint64 `json:"elapsedGameLoops"`
	Duration         int64  `json:"durationNanoseconds"`
	// UseScaledTime    bool          `json:"useScaledTime"`
	Version s2prot.Struct `json:"version"`
}
