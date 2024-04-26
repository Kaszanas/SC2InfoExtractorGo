package datastruct

// LogLevelEnum is an enum type which holds the available log levels.
type LogLevelEnum int

// Log levels:
const (
	Panic LogLevelEnum = iota
	Fatal
	Error
	Warn
	Info
	Debug
	Trace
)
