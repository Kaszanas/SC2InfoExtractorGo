package dataproc

import (
	"errors"
	"fmt"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils"
	log "github.com/sirupsen/logrus"
)

// ProcessReplayToJSON runs the processing pipeline on a single replay file
// and returns its JSON representation. Unlike PipelineWrapper it creates no
// progress bar, packages or processing info files.
func ProcessReplayToJSON(
	replayFile string,
	foreignToEnglishMapping map[string]string,
	cliFlags utils.CLIFlags,
) (string, error) {

	log.Debug("Entered ProcessReplayToJSON()")

	grpcAnonymizer := checkAnonymizationInitializeGRPC(
		cliFlags.PerformPlayerAnonymization,
	)
	if grpcAnonymizer != nil {
		defer func() {
			if err := grpcAnonymizer.Connection.Close(); err != nil {
				log.WithField("error", err).Error("Failed to close gRPC connection.")
			}
		}()
	}

	didWork, cleanReplayStructure, _, failureReason := FileProcessingPipeline(
		replayFile,
		grpcAnonymizer,
		foreignToEnglishMapping,
		cliFlags,
	)
	if !didWork {
		return "", fmt.Errorf("processing %s: %s", replayFile, failureReason)
	}

	stringifyOk, replayString := stringifyReplay(&cleanReplayStructure)
	if !stringifyOk {
		return "", errors.New("failed to convert the processed replay to JSON")
	}

	log.Debug("Finished ProcessReplayToJSON()")
	return replayString, nil
}
