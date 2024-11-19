package persistent_data

import "testing"

// TODO: Define processing info tests
// TestCollapseMapToMap tests the function collapseMapToMap that merges two maps.
func TestCollapseMapToMap(t *testing.T) {

	testCases := []struct {
		mapToCollapse map[string]int64
		collapseInto  map[string]int64
		expected      map[string]int64
	}{
		{
			map[string]int64{"c": 2},
			map[string]int64{"a": 1, "b": 2},
			map[string]int64{"a": 1, "b": 2, "c": 2},
		},
		{
			map[string]int64{"a": 1, "b": 2},
			map[string]int64{"c": 2},
			map[string]int64{"a": 1, "b": 2, "c": 2},
		},
		{
			map[string]int64{"a": 1, "b": 2},
			map[string]int64{"a": 1, "b": 2},
			map[string]int64{"a": 2, "b": 4},
		},
		{
			map[string]int64{"a": 1, "b": 2},
			map[string]int64{"a": 1, "b": 2, "c": 2},
			map[string]int64{"a": 2, "b": 4, "c": 2},
		},
		{
			map[string]int64{},
			map[string]int64{"a": 1, "b": 2, "c": 2},
			map[string]int64{"a": 1, "b": 2, "c": 2},
		},
		{
			map[string]int64{},
			map[string]int64{},
			map[string]int64{},
		},
	}

	for _, testCase := range testCases {
		collapseMapToMap(&testCase.mapToCollapse, &testCase.collapseInto)
		if !mapsAreEqual(testCase.collapseInto, testCase.expected) {
			t.Errorf("Expected %v, got %v", testCase.expected, testCase.collapseInto)
		}
	}

}

// TestAddReplaySummToPackageSumm tests the function that is adding
// a single replay summary to a package summary.
func TestAddReplaySummToPackageSumm(t *testing.T) {

	replaySummary := ReplaySummary{
		Summary: Summary{
			GameVersions: map[string]int64{"2.0.0": 1},
			GameTimes:    map[string]int64{"2": 1},
			Maps:         map[string]int64{"map1": 1},
			MapsGameTimes: GameTimes{
				map[string]map[string]int64{"map1": {"2": 1}}},
			Races:          map[string]int64{"Terran": 2},
			Units:          map[string]int64{"SCV": 2},
			OtherUnits:     map[string]int64{"Marine": 1},
			Dates:          map[string]int64{"2017-01-01": 1},
			DatesGameTimes: GameTimes{map[string]map[string]int64{"2017-01-01": {"2": 1}}},
			Servers:        map[string]int64{"EU": 1},
			MatchupCount:   map[string]int64{"TvT": 1},
			MatchupGameTimes: MatchupGameTimes{
				PvPMatchup: map[string]int64{},
				TvTMatchup: map[string]int64{"2": 1},
				ZvZMatchup: map[string]int64{},
				PvZMatchup: map[string]int64{},
				PvTMatchup: map[string]int64{},
				TvZMatchup: map[string]int64{},
			},
		}}

	packageSummary := PackageSummary{
		Summary: Summary{
			GameVersions: map[string]int64{"1.0.0": 1},
			GameTimes:    map[string]int64{"2": 1},
			Maps:         map[string]int64{"map1": 1},
			MapsGameTimes: GameTimes{
				map[string]map[string]int64{"map1": {"2": 1}}},
			Races:          map[string]int64{"Protoss": 1, "Terran": 1},
			Units:          map[string]int64{"Probe": 1, "SCV": 1},
			OtherUnits:     map[string]int64{"Zealot": 1},
			Dates:          map[string]int64{"2017-01-01": 1},
			DatesGameTimes: GameTimes{map[string]map[string]int64{"2017-01-01": {"2": 1}}},
			Servers:        map[string]int64{"EU": 1},
			MatchupCount:   map[string]int64{"PvT": 1},
			MatchupGameTimes: MatchupGameTimes{
				PvPMatchup: map[string]int64{},
				TvTMatchup: map[string]int64{},
				ZvZMatchup: map[string]int64{},
				PvZMatchup: map[string]int64{},
				PvTMatchup: map[string]int64{"2": 1},
				TvZMatchup: map[string]int64{},
			},
		}}

	AddReplaySummToPackageSumm(&replaySummary, &packageSummary)

	if len(packageSummary.Summary.GameVersions) != 2 {
		t.Errorf("Expected 2, got %v", len(packageSummary.Summary.GameVersions))
	}

	if len(packageSummary.Summary.GameTimes) != 1 {
		t.Errorf("Expected 1, got %v", len(packageSummary.Summary.GameTimes))
	}

	if len(packageSummary.Summary.MatchupCount) != 2 {
		t.Errorf("Expected 2, got %v", len(packageSummary.Summary.MatchupCount))
	}

}

// mapsAreEqual is a helper function that compares two maps to see if they are the same.
func mapsAreEqual(map1, map2 map[string]int64) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key, value := range map1 {
		if map2Value, ok := map2[key]; !ok || map2Value != value {
			return false
		}
	}

	return true
}
