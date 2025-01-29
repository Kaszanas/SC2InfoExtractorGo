package cleanup

import (
	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanTrackerEvents copies the tracker events,
// has the capability of removing unescessary fields.
func CleanTrackerEvents(replayData *rep.Rep) []s2prot.Struct {
	// Constructing a clean TrackerEvents without unescessary fields:
	var trackerEventsStructs []s2prot.Struct
	for _, trackerEvent := range replayData.TrackerEvts.Evts {

		// https://github.com/Kaszanas/SC2InfoExtractorGo/issues/41
		if trackerEvent.Struct["evtTypeName"] == "PlayerStats" {

			// Get stats:
			stats := trackerEvent.Struct["stats"].(s2prot.Struct)

			// Get values:
			foodUsed := stats["scoreValueFoodUsed"].(int64) / 4096
			foodMade := stats["scoreValueFoodMade"].(int64) / 4096

			// Overwrite values:
			trackerEvent.Struct["stats"].(s2prot.Struct)["scoreValueFoodUsed"] = foodUsed
			trackerEvent.Struct["stats"].(s2prot.Struct)["scoreValueFoodMade"] = foodMade
		}

		trackerEventsStructs = append(trackerEventsStructs, trackerEvent.Struct)
	}
	log.Info("Defined cleanTrackerEvents struct")
	return trackerEventsStructs
}
