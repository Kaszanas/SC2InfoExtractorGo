package dataproc

import (
	"fmt"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
	log "github.com/sirupsen/logrus"
)

func summarizeReplay(replayData *data.CleanedReplay) (bool, data.ReplaySummary) {

	log.Info("Entered summarizeReplay()")

	successFlag := true

	initSummary := data.DefaultReplaySummary()

	generateReplaySummary(replayData, &initSummary)

	fmt.Println(initSummary.Summary.MatchupHistograms, initSummary.Summary.Races)

	log.Info("Finished summarizeReplay()")
	return successFlag, initSummary

}
