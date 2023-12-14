package loglevel

// Log Levels that can be used
const (
	INFO = "info" // Info is the default log level. It is used for general information
	WARN = "warn" // Warn is used for errors that can be safely ignored
)

// All allowed log levels
var LogLevels = []string{INFO, WARN}
