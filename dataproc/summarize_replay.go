package dataproc

import (
	data "github.com/Kaszanas/GoSC2Science/datastruct"
)

// TODO: return a cleaned replay structure and stringify it elsewhere:

// This should be a place only for pipeline that performs cleaning on the incominng data.

// There should be another pipeline that is performing the summary update and returns a final replay string to be added to the zip archive.

func summarizeReplay(replayData *data.CleanedReplay) (bool, data.ReplaySummary) {

	successFlag := true

	initSummaryStruct := data.ReplaySummary{}

	generateReplaySummary(*replayData, &initSummaryStruct)

	return successFlag, initSummaryStruct

}
