package settings

// UnusedMessageEvents returns a slice of event names that are to be excluded when obtaining final replay data.
var UnusedMessageEvents = []string{
	"LoadingProgress",
}

// AnonymizeMessageEvents is a slice of events that are going to be deleted from the replay structures for the need of anonymization
var AnonymizeMessageEvents = []string{
	"Chat",
}
