package cleanup

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanDetails copies the details,
// has the capability of removing unescessary fields.
func CleanDetails(replayData *rep.Rep) (replay_data.CleanedDetails, replay_data.ReplayMapField) {
	// Constructing a clean CleanedDetails without unescessary fields
	detailsGameSpeed := replayData.Details.GameSpeed().String()
	detailsIsBlizzardMap := replayData.Details.IsBlizzardMap()

	// mapFileName := replayData.Details.MapFileName()
	// log.WithField("mapFileName", mapFileName).Info("Found mapFileName")

	timeUTC := replayData.Details.TimeUTC()
	mapNameString := replayData.Details.Title()
	replayMapField := replay_data.ReplayMapField{
		MapName: mapNameString,
	}

	cleanDetails := replay_data.CleanedDetails{
		GameSpeed:     detailsGameSpeed,
		IsBlizzardMap: detailsIsBlizzardMap,
		// PlayerList:    detailsPlayerList, // Information from that part is merged with ToonDescMap
		// TimeLocalOffset: timeLocalOffset, // This is unused
		TimeUTC: timeUTC,
		// MapName: mapNameString, // This is unused
	}
	log.Info("Defined cleanDetails struct")
	return cleanDetails, replayMapField
}
