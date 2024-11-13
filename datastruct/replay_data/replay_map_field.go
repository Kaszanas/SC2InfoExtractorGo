package replay_data

type ReplayMapField struct {
	MapName string
}

// CombineReplayMapFields Detects first non-empty map name from the list of ReplayMapFields
func CombineReplayMapFields(rmfs []ReplayMapField) string {

	mapName := ""
	for _, rmf := range rmfs {
		if rmf.MapName != "" {
			mapName = rmf.MapName
			break
		}
	}

	return mapName
}
