package dataproc

import (
	data "github.com/Kaszanas/SC2InfoExtractorGo/datastruct"
	log "github.com/sirupsen/logrus"
)

// summarizeReplay accesses information from within a replay and creates histograms, counters etc. in order to visualize the replay contents.
func summarizeReplay(replayData *data.CleanedReplay) (bool, data.ReplaySummary) {

	log.Info("Entered summarizeReplay()")

	successFlag := true

	initSummary := data.DefaultReplaySummary()

	generateReplaySummary(replayData, &initSummary)

	// fmt.Println(initSummary.Summary.MatchupHistograms, initSummary.Summary.Races)

	log.Info("Finished summarizeReplay()")
	return successFlag, initSummary

}
