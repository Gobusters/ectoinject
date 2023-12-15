package logging

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Gobusters/ectoinject/loglevel"
)

// STDOUT colors
const (
	reset  = "\033[0m"
	yellow = "\033[33m"
	blue   = "\033[34m"
)

type LogFunc func(ctx context.Context, level, msg string)

type Logger struct {
	prefix         string
	level          string
	colorsEnabled  bool
	customLogFunc  LogFunc
	loggingEnabled bool
}

// NewLogger Creates a new Logger
// prefix: The prefix to use for all log messages. Defaults to "ectoinject"
// level: The log level to use. Must be one of INFO, WARN, ERROR, FATAL (defaults to INFO)
// colorsEnabled: Whether or not to use colors in the log messages
// loggingEnabled: Whether or not to log messages
// customLogFunc: A custom log function to use. If provided, all other options are ignored
func NewLogger(prefix, level string, colorsEnabled, loggingEnabled bool, customLogFunc LogFunc) (*Logger, error) {
	level = strings.ToLower(level)
	if level == "" {
		level = loglevel.INFO
	}

	if prefix == "" {
		prefix = "ectoinject"
	}

	// Ensure the log level is valid
	if !validateLogLevel(level) {
		return nil, fmt.Errorf("invalid log level '%s' must be one of %v", level, loglevel.LogLevels)
	}

	return &Logger{
		prefix:         prefix,
		level:          level,
		colorsEnabled:  colorsEnabled,
		customLogFunc:  customLogFunc,
		loggingEnabled: loggingEnabled,
	}, nil
}

func validateLogLevel(level string) bool {
	// Ensure the log level is valid
	valid := false
	for _, l := range loglevel.LogLevels {
		if l == level {
			valid = true
			break
		}
	}

	return valid
}

// Warn Logs a message at the WARN level
// format: The format string to use
// args: The arguments to use in the format string
func (l *Logger) Warn(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.LogMessage(ctx, loglevel.WARN, msg)
}

// Info Logs a message at the INFO level
// format: The format string to use
// args: The arguments to use in the format string
func (l *Logger) Info(ctx context.Context, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.LogMessage(ctx, loglevel.INFO, msg)
}

// LogMessage Logs a message to STDOUT
// level: The log level to use
// msg: The message to log
func (l *Logger) LogMessage(ctx context.Context, level, msg string) {
	// Are logs enabled?
	if !l.loggingEnabled {
		return
	}

	// Use custom log function if provided
	if l.customLogFunc != nil {
		l.customLogFunc(ctx, level, msg)
		return
	}

	// Adds log info to the message
	msg = fmt.Sprintf("%s (%s): %s", l.prefix, level, msg)

	color := ""
	resetColor := ""

	if l.colorsEnabled {
		resetColor = reset

		switch level {
		case loglevel.INFO:
			color = blue
		case loglevel.WARN:
			color = yellow
		}
	}

	if level == loglevel.WARN {
		log.Printf("%s%s%s", color, msg, resetColor)
		return
	}

	if l.level == loglevel.WARN {
		return // Don't log anything under WARN
	}

	if level == loglevel.INFO {
		log.Printf("%s%s%s", color, msg, resetColor)
		return
	}

	log.Println(msg)
}
