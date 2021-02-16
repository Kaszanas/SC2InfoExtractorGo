package settings

// UnusedMessageEvents returns a slice of event names that are to be excluded when obtaining final replay data.
var UnusedMessageEvents = []string{
	"LoadingProgress",
	"Chat",
}
