package main

import (
	"time"

	"github.com/icza/s2prot"
)

type CleanedHeader struct {
	elapsedGameLoops uint64
	duration         time.Duration
	useScaledTime    bool
	version          s2prot.Struct
}
