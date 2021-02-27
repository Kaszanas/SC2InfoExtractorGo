package datastruct

// PackageSummary contains statistics calculated from replay information that belong to a whole ZIP archive.
type PackageSummary struct {
	Summary Summary
}

// ReplaySummary contains information calculated from a single replay
type ReplaySummary struct {
	Summary Summary
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
	PvPMatchup map[int64]int64 `json:"PvPMatchup"`
	TvTMatchup map[int64]int64 `json:"TvTMatchup"`
	ZvZMatchup map[int64]int64 `json:"ZvZMatchup"`
	PvZMatchup map[int64]int64 `json:"PvZMatchup"`
	PvTMatchup map[int64]int64 `json:"PvTMatchup"`
	TvZMatchup map[int64]int64 `json:"TvZMatchup"`
}

// DefaultMatchupTime ...
func DefaultMatchupTime() MatchupTime {

	return MatchupTime{
		Matchup:   make(map[string]int64),
		GameTimes: make(map[string]int64),
	}

}
