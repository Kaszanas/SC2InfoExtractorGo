package datastruct

// PackageSummary contains statistics calculated from replay information that belong to a whole ZIP archive.
type PackageSummary struct {
	Summary Summary
}

// ReplaySummary contains information calculated from a single replay
type ReplaySummary struct {
	Summary Summary
}

// AddReplaySummToPackageSumm adds the replay summary to the package summary.
func AddReplaySummToPackageSumm(packageSummary *PackageSummary, replaySummary *ReplaySummary) {

	// Adding GameVersion information to PackageSummary
	replayGameVersions := replaySummary.Summary.GameVersions
	packageGameVersions := packageSummary.Summary.GameVersions
	collapsedGameVersions := collapseMapToMap(replayGameVersions, packageGameVersions)
	packageSummary.Summary.GameVersions = collapsedGameVersions

	// Adding GameTimes information to PackageSummary
	replayGameTimes := replaySummary.Summary.GameTimes
	packageGameTimes := packageSummary.Summary.GameTimes
	collapsedGameTimes := collapseMapToMap(replayGameTimes, packageGameTimes)
	packageSummary.Summary.GameTimes = collapsedGameTimes

	// Adding Maps information to PackageSummary
	replayMaps := replaySummary.Summary.Maps
	packageMaps := replaySummary.Summary.Maps
	collapsedMaps := collapseMapToMap(replayMaps, packageMaps)
	packageSummary.Summary.Maps = collapsedMaps

	// Adding Races information to PackageSummary
	replayRaces := replaySummary.Summary.Races
	packageRaces := packageSummary.Summary.Races
	collapsedRaces := collapseMapToMap(replayRaces, packageRaces)
	packageSummary.Summary.Races = collapsedRaces

	// Adding Units information to PackageSummary
	replayUnits := replaySummary.Summary.Units
	packageUnits := packageSummary.Summary.Units
	collapsedUnits := collapseMapToMap(replayUnits, packageUnits)
	packageSummary.Summary.Units = collapsedUnits

	// Adding Dates information to PackageSummary
	replayDates := replaySummary.Summary.Dates
	packageDates := packageSummary.Summary.Dates
	collapsedDates := collapseMapToMap(replayDates, packageDates)
	packageSummary.Summary.Dates = collapsedDates

	// Adding Servers information to PackageSummary
	replayServers := replaySummary.Summary.Servers
	packageServers := packageSummary.Summary.Servers
	collapsedServers := collapseMapToMap(replayServers, packageServers)
	packageSummary.Summary.Servers = collapsedServers

	// Adding matchup information to the PackageSummary
	replayPvP := replaySummary.Summary.MatchupHistograms.PvPMatchup
	packageSummary.Summary.MatchupHistograms.PvPMatchup = packageSummary.Summary.MatchupHistograms.PvPMatchup + replayPvP

	replayTvT := replaySummary.Summary.MatchupHistograms.TvTMatchup
	packageSummary.Summary.MatchupHistograms.TvTMatchup = packageSummary.Summary.MatchupHistograms.TvTMatchup + replayTvT

	replayZvZ := replaySummary.Summary.MatchupHistograms.ZvZMatchup
	packageSummary.Summary.MatchupHistograms.ZvZMatchup = packageSummary.Summary.MatchupHistograms.ZvZMatchup + replayZvZ

	replayPvZ := replaySummary.Summary.MatchupHistograms.PvZMatchup
	packageSummary.Summary.MatchupHistograms.PvZMatchup = packageSummary.Summary.MatchupHistograms.PvZMatchup + replayPvZ

	replayPvT := replaySummary.Summary.MatchupHistograms.PvTMatchup
	packageSummary.Summary.MatchupHistograms.PvTMatchup = packageSummary.Summary.MatchupHistograms.PvTMatchup + replayPvT

	replayTvZ := replaySummary.Summary.MatchupHistograms.TvZMatchup
	packageSummary.Summary.MatchupHistograms.TvTMatchup = packageSummary.Summary.MatchupHistograms.TvTMatchup + replayTvZ

}

// collapseMapToMap adds the keys and values of one map to another.
func collapseMapToMap(mapToCollapse map[string]int64, collapseInto map[string]int64) map[string]int64 {

	for key, value := range mapToCollapse {
		collapseValue, ok := collapseInto[key]
		if ok {
			collapseInto[key] = collapseValue + value
		} else {
			collapseInto[key] = value
		}
	}

	return collapseInto
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
	GameVersions      map[string]int64  `json:"gameVersions"`
	GameTimes         map[string]int64  `json:"gameTimes"`
	Maps              map[string]int64  `json:"maps"`
	Races             map[string]int64  `json:"races"`
	Units             map[string]int64  `json:"units"`
	Dates             map[string]int64  `json:"dates"`
	Servers           map[string]int64  `json:"servers"`
	MatchupHistograms MatchupHistograms `json:"matchupHistograms"`
}

// DefaultSummary ...
func DefaultSummary() Summary {

	return Summary{
		GameVersions: make(map[string]int64),
		GameTimes:    make(map[string]int64),
		Maps:         make(map[string]int64),
		Races:        make(map[string]int64),
		Units:        make(map[string]int64),
		Dates:        make(map[string]int64),
		Servers:      make(map[string]int64),
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
