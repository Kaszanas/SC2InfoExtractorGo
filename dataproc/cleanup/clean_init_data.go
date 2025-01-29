package cleanup

import (
	"github.com/Kaszanas/SC2InfoExtractorGo/datastruct/replay_data"
	"github.com/icza/s2prot/rep"
	log "github.com/sirupsen/logrus"
)

// cleanInitData copies the init data,
// partly verifies the integrity of the data.
// Has the capability to remove unescessary fields.
func CleanInitData(
	replayData *rep.Rep,
	cleanedGameDescription replay_data.CleanedGameDescription) (
	replay_data.CleanedInitData,
	[]replay_data.CleanedUserInitData,
	bool) {
	// Constructing a clean UserInitData without unescessary fields:
	var cleanedUserInitDataList []replay_data.CleanedUserInitData
	for _, userInitData := range replayData.InitData.UserInitDatas {
		// If the name is an empty string ommit the struct and enter next iteration:
		name := userInitData.Name()
		if !(len(name) > 0) {
			continue
		}

		combinedRaceLevels := userInitData.CombinedRaceLevels()
		highestLeague := userInitData.HighestLeague().String()
		clanTag := userInitData.ClanTag()
		isInClan := checkClan(clanTag)

		userInitDataStruct := replay_data.CleanedUserInitData{
			CombinedRaceLevels: uint64(combinedRaceLevels),
			HighestLeague:      highestLeague,
			Name:               name,
			IsInClan:           isInClan,
			ClanTag:            clanTag,
		}

		cleanedUserInitDataList = append(
			cleanedUserInitDataList,
			userInitDataStruct,
		)
	}

	cleanInitData := replay_data.CleanedInitData{
		GameDescription: cleanedGameDescription,
	}
	log.Info("Defined cleanInitData struct")
	return cleanInitData, cleanedUserInitDataList, true
}
