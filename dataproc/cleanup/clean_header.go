package cleanup

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanHeader copies the header,
// has the capability of removing unescessary fields.
func CleanHeader(replayData *rep.Rep) replay_data.CleanedHeader {
	// Constructing a clean replay header without unescessary fields:
	elapsedGameLoops := replayData.Header.Loops()
	// TODO: These values of duration are not verified: https://github.com/icza/s2prot/issues/48
	// durationNanoseconds := replayData.Header.Duration().Nanoseconds()
	// durationSeconds := replayData.Header.Duration().Seconds()
	// version := replayData.Header.Struct["version"].(s2prot.Struct)

	version := replayData.Header.VersionString()

	cleanHeader := replay_data.CleanedHeader{
		ElapsedGameLoops: uint64(elapsedGameLoops),
		// DurationNanoseconds: durationNanoseconds,
		// DurationSeconds:     durationSeconds,
		Version: version,
	}
	log.Info("Defined cleanHeader struct")
	return cleanHeader
}
