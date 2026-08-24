package charmlogger

import (
	"io"
	"log/slog"

	charmlog "charm.land/log/v2"
)

var handler *charmlog.Logger

// Setup configures the global logger with the specified level and format.
// Valid log levels are: debug, info, warn, error, fatal.
// If an invalid level is provided, it defaults to info.
// When jsonFormat is true, logs are output as JSON for machine parsing.
func Setup(level string, destination io.Writer, jsonFormat bool) *slog.Logger {
	handler = charmlog.New(destination)
	SetLevel(level)

	handler.SetReportTimestamp(true)

	if jsonFormat {
		handler.SetFormatter(charmlog.JSONFormatter)
	} else {
		handler.SetFormatter(charmlog.TextFormatter)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func SetLevel(level string) {
	lvl, err := charmlog.ParseLevel(level)
	if err != nil {
		lvl = charmlog.InfoLevel
	}
	handler.SetLevel(lvl)
}

// GetLogger returns the global logger instance.
// The logger should be configured via Setup before use.
func GetLogger() *slog.Logger {
	return slog.Default()
}
