package datastruct

// PackageSummary contains statistics calculated from replay information
type PackageSummary struct {
	GameVersions      map[string]int64  `json:"gameVersions"`
	GameTimes         map[string]int64  `json:"gameTimes"`
	Maps              map[string]int64  `json:"maps"`
	Races             map[string]int64  `json:"races"`
	Units             map[string]int64  `json:"units"`
	Dates             map[string]int64  `json:"dates"`
	Servers           map[string]int64  `json:"servers"`
	MatchupHistograms MatchupHistograms `json:"matchupHistograms"`
}

// MatchupHistograms aggregates the data that is required to prepare histograms of Matchup vs Game Length
type MatchupHistograms struct {
	PvPMatchup []MatchupTime `json:"PvPMatchup"`
	TvTMatchup []MatchupTime `json:"TvTMatchup"`
	ZvZMatchup []MatchupTime `json:"ZvZMatchup"`
	PvZMatchup []MatchupTime `json:"PvZMatchup"`
	PvTMatchup []MatchupTime `json:"PvTMatchup"`
	TvZMatchup []MatchupTime `json:"TvZMatchup"`
}

// MatchupTime contains information about game length vs the current matchup
type MatchupTime struct {
	// TODO: This design is not sufficient and does not fit the data that is required:
	Matchup   map[string]int64 `json:"matchup"`
	GameTimes map[string]int64 `json:"gameTimes"`
}
