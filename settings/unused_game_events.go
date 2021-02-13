package settings

// UnusedGameEvents returns a slice of event names that are to be excluded when obtaining final replay data.
func UnusedGameEvents() []string {

	unusedEvents := []string{
		"TriggerSoundLengthSync",
		"SetSyncPlayingTime",
	}

	return unusedEvents
}
