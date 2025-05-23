package persistent_data

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/Kaszanas/SC2InfoExtractorGo/utils/file_utils"
	log "github.com/sirupsen/logrus"
)

// PackageSummary is a structure contains statistics
// calculated from replay information that belong to a whole ZIP archive.
type PackageSummary struct {
	Summary Summary
}

// ReplaySummary contains information calculated from a single replay
type ReplaySummary struct {
	Summary Summary
}

// NewPackageSummary returns an initialized PackageSummary
func NewPackageSummary() PackageSummary {
	return PackageSummary{Summary: NewSummary()}
}

// NewReplaySummary returns an initialized ReplaySummary
func NewReplaySummary() ReplaySummary {
	return ReplaySummary{Summary: NewSummary()}
}

// Summary is an abstract type used by both ReplaySummary
// and PackageSummary and contains fields that are used as descriptive statistics
type Summary struct {
	GameVersions     map[string]int64 `json:"gameVersions"`
	GameTimes        map[string]int64 `json:"gameTimes"`
	Maps             map[string]int64 `json:"maps"`
	MapsGameTimes    GameTimes        `json:"mapGameTimes"`
	Races            map[string]int64 `json:"races"`
	Units            map[string]int64 `json:"units"`
	OtherUnits       map[string]int64 `json:"otherUnits"`
	Dates            map[string]int64 `json:"dates"`
	DatesGameTimes   GameTimes        `json:"datesGameTimes"`
	Servers          map[string]int64 `json:"servers"`
	MatchupCount     map[string]int64 `json:"matchupCount"`
	MatchupGameTimes MatchupGameTimes `json:"matchupGameTimes"`
}

// NewSummary returns a Summary structure with initialized fiends
func NewSummary() Summary {

	return Summary{
		GameVersions:     make(map[string]int64),
		GameTimes:        make(map[string]int64),
		Maps:             make(map[string]int64),
		MapsGameTimes:    NewGameTimes(),
		Races:            make(map[string]int64),
		Units:            make(map[string]int64),
		OtherUnits:       make(map[string]int64),
		Dates:            make(map[string]int64),
		DatesGameTimes:   NewGameTimes(),
		Servers:          make(map[string]int64),
		MatchupCount:     make(map[string]int64),
		MatchupGameTimes: NewMatchupGameTimes(),
	}
}

// MatchupHistograms aggregates the data that is required
// to prepare histograms of Matchup vs Game Length
type MatchupGameTimes struct {
	PvPMatchup map[string]int64 `json:"PvPMatchupGameTimes"`
	TvTMatchup map[string]int64 `json:"TvTMatchupGameTimes"`
	ZvZMatchup map[string]int64 `json:"ZvZMatchupGameTimes"`
	PvZMatchup map[string]int64 `json:"PvZMatchupGameTimes"`
	PvTMatchup map[string]int64 `json:"PvTMatchupGameTimes"`
	TvZMatchup map[string]int64 `json:"TvZMatchupGameTimes"`
}

// NewMatchupHistograms returns a structure with initialized fields.
func NewMatchupGameTimes() MatchupGameTimes {

	return MatchupGameTimes{
		PvPMatchup: make(map[string]int64),
		TvTMatchup: make(map[string]int64),
		ZvZMatchup: make(map[string]int64),
		PvZMatchup: make(map[string]int64),
		PvTMatchup: make(map[string]int64),
		TvZMatchup: make(map[string]int64),
	}

}

type GameTimes struct {
	GameTimes map[string]map[string]int64 `json:"gameTimes"`
}

func NewGameTimes() GameTimes {
	return GameTimes{
		GameTimes: make(map[string]map[string]int64),
	}
}

// CreatePackageSummaryFile receives packageSummaryStruct and fileNumber
// then saves the package summary file onto the drive.
func CreatePackageSummaryFile(
	absolutePathOutputDirectory string,
	packageSummaryStruct PackageSummary,
	fileNumber int,
) error {
	log.Debug("Entered CreatePackageSummaryFile()")

	packageSummaryFilename := fmt.Sprintf("package_summary_%v.json", fileNumber)
	packageAbsolutePath := filepath.Join(absolutePathOutputDirectory, packageSummaryFilename)
	packageSummaryFile, err := file_utils.CreateTruncateFile(packageAbsolutePath)
	if err != nil {
		log.Error("Failed to create the package summary file!")
		return err
	}

	packageSummaryBytes, err := json.Marshal(packageSummaryStruct)
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to marshal packageSummaryStruct")
		return fmt.Errorf("Failed to marshal packageSummaryStruct: %v", err)
	}
	_, err = packageSummaryFile.Write(packageSummaryBytes)
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to save the packageSummaryFile")
		return fmt.Errorf("Failed to save the packageSummaryFile: %v", err)
	}

	err = packageSummaryFile.Close()
	if err != nil {
		log.WithField("error", err).
			Fatal("Failed to cloes the packageSummaryFile")
		return fmt.Errorf("Failed to close the packageSummaryFile: %v", err)
	}

	log.Debug("Finished CreatePackageSummaryFile()")
	return nil
}

// AddReplaySummToPackageSumm adds the replay summary to the package summary.
func AddReplaySummToPackageSumm(
	replaySummary *ReplaySummary,
	packageSummary *PackageSummary,
) {

	log.Debug("Entered AddReplaySummToPackageSumm()")

	// Adding GameVersion information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.GameVersions,
		&packageSummary.Summary.GameVersions)
	log.Info("Finished collapsing GameVersions")

	// Adding GameTimes information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.GameTimes,
		&packageSummary.Summary.GameTimes)
	log.Info("Finished collapsing GameTimes")

	// Adding Maps information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.Maps,
		&packageSummary.Summary.Maps)
	log.Info("Finished collapsing Maps")

	// Adding Races information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.Races,
		&packageSummary.Summary.Races)
	log.Info("Finished collapsing Races")

	// Adding Units information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.Units,
		&packageSummary.Summary.Units)
	log.Info("Finished collapsing Units")
	collapseMapToMap(
		&replaySummary.Summary.OtherUnits,
		&packageSummary.Summary.OtherUnits)
	log.Info("Finished collapsing OtherUnits")

	// Adding Dates information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.Dates,
		&packageSummary.Summary.Dates)
	log.Info("Finished collapsing Dates")

	// Creating nested structures for game times by dates:
	for key, replayGameTimeMap := range replaySummary.Summary.DatesGameTimes.GameTimes {
		if packageSummaryMap, ok := packageSummary.Summary.DatesGameTimes.GameTimes[key]; ok {
			collapseMapToMap(&replayGameTimeMap, &packageSummaryMap)
			packageSummary.Summary.DatesGameTimes.GameTimes[key] = packageSummaryMap
		} else {
			packageSummary.Summary.DatesGameTimes.GameTimes[key] = replayGameTimeMap
		}
	}

	// Creating nested structures for game times by maps:
	for key, replayGameTimeMap := range replaySummary.Summary.MapsGameTimes.GameTimes {
		if packageSummaryMap, ok := packageSummary.Summary.MapsGameTimes.GameTimes[key]; ok {
			collapseMapToMap(&replayGameTimeMap, &packageSummaryMap)
			packageSummary.Summary.MapsGameTimes.GameTimes[key] = packageSummaryMap
		} else {
			packageSummary.Summary.MapsGameTimes.GameTimes[key] = replayGameTimeMap
		}
	}

	// Adding Servers information to PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.Servers,
		&packageSummary.Summary.Servers)
	log.Info("Finished collapsing Servers")

	// Adding matchup count information to the PackageSummary:
	collapseMapToMap(
		&replaySummary.Summary.MatchupCount,
		&packageSummary.Summary.MatchupCount)

	// Collapsing all of the matchup game times:
	collapseMapToMap(
		&replaySummary.Summary.MatchupGameTimes.PvPMatchup,
		&packageSummary.Summary.MatchupGameTimes.PvPMatchup)
	collapseMapToMap(
		&replaySummary.Summary.MatchupGameTimes.PvTMatchup,
		&packageSummary.Summary.MatchupGameTimes.PvTMatchup)
	collapseMapToMap(
		&replaySummary.Summary.MatchupGameTimes.PvZMatchup,
		&packageSummary.Summary.MatchupGameTimes.PvZMatchup)
	collapseMapToMap(
		&replaySummary.Summary.MatchupGameTimes.TvTMatchup,
		&packageSummary.Summary.MatchupGameTimes.TvTMatchup)
	collapseMapToMap(
		&replaySummary.Summary.MatchupGameTimes.TvZMatchup,
		&packageSummary.Summary.MatchupGameTimes.TvZMatchup)
	collapseMapToMap(
		&replaySummary.Summary.MatchupGameTimes.ZvZMatchup,
		&packageSummary.Summary.MatchupGameTimes.ZvZMatchup)

	log.Info("Finished collapsing matchup information")
	log.Debug("Finished AddReplaySummToPackageSumm()")
}

// collapseMapToMap adds the keys and values of one map to another.
func collapseMapToMap(
	mapToCollapse *map[string]int64,
	collapseInto *map[string]int64,
) {

	log.Debug("Entered collapseMapToMap()")

	for key, value := range *mapToCollapse {
		collapseValue, ok := (*collapseInto)[key]
		if ok {
			(*collapseInto)[key] = collapseValue + value
		} else {
			(*collapseInto)[key] = value
		}
	}

	log.Debug("Finished collapseMapToMap()")
}
