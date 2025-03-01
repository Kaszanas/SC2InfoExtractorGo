package dataproc

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/persistent_data"
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	log "github.com/sirupsen/logrus"
)

// summarizeReplay accesses information from within a replay
// and creates histograms, counters etc. in order to visualize the replay contents.
func summarizeReplay(replayData *replay_data.CleanedReplay) (bool, persistent_data.ReplaySummary) {

	log.Debug("Entered summarizeReplay()")

	initSummary := persistent_data.NewReplaySummary()

	generateReplaySummary(replayData, &initSummary)

	log.Debug("Finished summarizeReplay()")
	return true, initSummary
}
