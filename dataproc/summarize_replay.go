package dataproc

import (
	"fmt"

	data "github.com/Kaszanas/GoSC2Science/datastruct"
)

func summarizeReplay(replayData *data.CleanedReplay) (bool, data.ReplaySummary) {

	successFlag := true

	initSummary := data.DefaultReplaySummary()

	generateReplaySummary(replayData, &initSummary)

	fmt.Println(initSummary)

	return successFlag, initSummary

}
