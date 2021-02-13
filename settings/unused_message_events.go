package settings

// UnusedMessageEvents returns a slice of event names that are to be excluded when obtaining final replay data.
func UnusedMessageEvents() []string {

	unusedEvents := []string{
		"LoadingProgress",
		"Chat",
	}

	return unusedEvents
}
