package loglevel

// Log Levels that can be used
const (
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
	FATAL = "fatal"
)

// All allowed log levels
var LogLevels = []string{INFO, WARN, ERROR, FATAL}
