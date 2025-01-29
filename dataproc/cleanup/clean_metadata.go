package cleanup

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanMetadata copies the metadata,
// has the capability of removing unescessary fields.
func CleanMetadata(
	replayData *rep.Rep,
) (replay_data.CleanedMetadata, replay_data.ReplayMapField) {
	// Constructing a clean CleanedMetadata without unescessary fields:
	metadataBaseBuild := replayData.Metadata.BaseBuild()
	metadataDataBuild := replayData.Metadata.DataBuild()
	// metadataDuration := replayData.Metadata.DurationSec()
	metadataGameVersion := replayData.Metadata.GameVersion()

	foreignMetadataMapName := replayData.Metadata.Title()
	mapNameField := replay_data.ReplayMapField{
		MapName: foreignMetadataMapName,
	}

	cleanMetadata := replay_data.CleanedMetadata{
		BaseBuild: metadataBaseBuild,
		DataBuild: metadataDataBuild,
		// Duration:    metadataDuration,
		GameVersion: metadataGameVersion,
		// Players:     metadataCleanedPlayersList, // This is unused.
		MapName: foreignMetadataMapName,
	}
	log.Info("Defined cleanMetadata struct")
	return cleanMetadata, mapNameField
}
