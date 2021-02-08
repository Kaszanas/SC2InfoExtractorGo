package datastruct

import (
	"time"

	"github.com/icza/s2prot"
)

type CleanedHeader struct {
	ElapsedGameLoops uint64
	Duration         time.Duration
	UseScaledTime    bool
	Version          s2prot.Struct
}
