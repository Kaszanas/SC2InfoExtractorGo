package settings

// UnusedGameEvents returns a slice of event names that are
// to be excluded when obtaining final replay data.
var UnusedGameEvents = []string{
	"TriggerSoundLengthSync",
	"SetSyncPlayingTime",
	"SetSyncLoadingTime",
	"UserFinishedLoadingSync",
}
