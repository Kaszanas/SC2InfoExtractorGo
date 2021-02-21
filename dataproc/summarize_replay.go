package dataproc

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
)

func summarizeReplay(replayData *data.CleanedReplay) (bool, data.ReplaySummary) {

	successFlag := true

	initSummaryStruct := data.ReplaySummary{}

	generateReplaySummary(*replayData, &initSummaryStruct)

	return successFlag, initSummaryStruct

}
