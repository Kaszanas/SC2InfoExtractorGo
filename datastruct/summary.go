package datastruct

import (
	log "github.com/sirupsen/logrus"
)

// AddReplaySummToPackageSumm adds the replay summary to the package summary.
func AddReplaySummToPackageSumm(replaySummary *ReplaySummary, packageSummary *PackageSummary) {

	log.Info("Entered AddReplaySummToPackageSumm()")

	// Adding GameVersion information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.GameVersions, &packageSummary.Summary.GameVersions)
	log.Info("Finished collapsing GameVersions")

	// Adding GameTimes information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.GameTimes, &packageSummary.Summary.GameTimes)
	log.Info("Finished collapsing GameTimes")

	// Adding Maps information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.Maps, &replaySummary.Summary.Maps)
	log.Info("Finished collapsing Maps")

	// Adding Races information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.Races, &packageSummary.Summary.Races)
	log.Info("Finished collapsing Races")

	// Adding Units information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.Units, &packageSummary.Summary.Units)
	log.Info("Finished collapsing Units")

	// Adding Dates information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.Dates, &packageSummary.Summary.Dates)
	log.Info("Finished collapsing Dates")

	// Adding Servers information to PackageSummary:
	collapseMapToMap(&replaySummary.Summary.Servers, &packageSummary.Summary.Servers)
	log.Info("Finished collapsing Servers")

	// Adding matchup information to the PackageSummary:
	// TODO: Check if this is working?
	collapseMapToMap(&replaySummary.Summary.MatchupHistograms, &packageSummary.Summary.MatchupHistograms)
	log.Info("Finished collapsing matchup information")

	// packageSummary.Summary.MatchupHistograms["PvP"] = packageSummary.Summary.MatchupHistograms["PvP"] + replaySummary.Summary.MatchupHistograms.PvPMatchup
	// log.Info("Finished collapsing PvPMatchup")

	// packageSummary.Summary.MatchupHistograms["TvT"] = packageSummary.Summary.MatchupHistograms.["TvT"] + replaySummary.Summary.MatchupHistograms.TvTMatchup
	// log.Info("Finished collapsing TvTMatchup")

	// packageSummary.Summary.MatchupHistograms["ZvZ"] = packageSummary.Summary.MatchupHistograms["ZvZ"] + replaySummary.Summary.MatchupHistograms.ZvZMatchup
	// log.Info("Finished collapsing ZvZMatchup")

	// packageSummary.Summary.MatchupHistograms["PvZ"] = packageSummary.Summary.MatchupHistograms.PvZMatchup + replaySummary.Summary.MatchupHistograms.PvZMatchup
	// log.Info("Finished collapsing PvZMatchup")

	// packageSummary.Summary.MatchupHistograms["PvT"] = packageSummary.Summary.MatchupHistograms.PvTMatchup + replaySummary.Summary.MatchupHistograms.PvTMatchup
	// log.Info("Finished collapsing PvTMatchup")

	// packageSummary.Summary.MatchupHistograms.TvTMatchup = packageSummary.Summary.MatchupHistograms.TvTMatchup + replaySummary.Summary.MatchupHistograms.TvZMatchup
	// log.Info("Finished collapsing PvTMatchup")

	log.Info("Finished AddReplaySummToPackageSumm()")

}

// collapseMapToMap adds the keys and values of one map to another.
func collapseMapToMap(mapToCollapse *map[string]int64, collapseInto *map[string]int64) {

	log.Info("Entered collapseMapToMap()")

	for key, value := range *mapToCollapse {
		collapseValue, ok := (*collapseInto)[key]
		if ok {
			(*collapseInto)[key] = collapseValue + value
		} else {
			(*collapseInto)[key] = value
		}
	}

	log.Info("Finished collapseMapToMap()")
}

// PackageSummary is a structure contains statistics calculated from replay information that belong to a whole ZIP archive.
type PackageSummary struct {
	Summary Summary
}

// ReplaySummary contains information calculated from a single replay
type ReplaySummary struct {
	Summary Summary
}

// DefaultPackageSummary returns an initialized PackageSummary
func DefaultPackageSummary() PackageSummary {
	return PackageSummary{Summary: DefaultSummary()}
}

// DefaultReplaySummary returns an initialized ReplaySummary
func DefaultReplaySummary() ReplaySummary {
	return ReplaySummary{Summary: DefaultSummary()}
}

// Summary is an abstract type used by both ReplaySummary and PackageSummary and contains fields that are used as descriptive statistics
type Summary struct {
	GameVersions      map[string]int64 `json:"gameVersions"`
	GameTimes         map[string]int64 `json:"gameTimes"`
	Maps              map[string]int64 `json:"maps"`
	Races             map[string]int64 `json:"races"`
	Units             map[string]int64 `json:"units"`
	Dates             map[string]int64 `json:"dates"`
	Servers           map[string]int64 `json:"servers"`
	MatchupHistograms map[string]int64 `json:"matchupHistograms"`
}

// DefaultSummary ...
func DefaultSummary() Summary {

	return Summary{
		GameVersions:      make(map[string]int64),
		GameTimes:         make(map[string]int64),
		Maps:              make(map[string]int64),
		Races:             make(map[string]int64),
		Units:             make(map[string]int64),
		Dates:             make(map[string]int64),
		Servers:           make(map[string]int64),
		MatchupHistograms: make(map[string]int64),
	}
}

// MatchupHistograms aggregates the data that is required to prepare histograms of Matchup vs Game Length
type MatchupHistograms struct {
	PvPMatchup int64 `json:"PvPMatchup"`
	TvTMatchup int64 `json:"TvTMatchup"`
	ZvZMatchup int64 `json:"ZvZMatchup"`
	PvZMatchup int64 `json:"PvZMatchup"`
	PvTMatchup int64 `json:"PvTMatchup"`
	TvZMatchup int64 `json:"TvZMatchup"`
}

// DefaultMatchupHistograms ...
// func DefaultMatchupHistograms() MatchupHistograms {

// 	return MatchupHistograms{
// 		PvPMatchup: make(int64),
// 		TvTMatchup: make(map[int64]int64),
// 		ZvZMatchup: make(map[int64]int64),
// 		PvZMatchup: make(map[int64]int64),
// 		PvTMatchup: make(map[int64]int64),
// 		TvZMatchup: make(map[int64]int64),
// 	}

// }
