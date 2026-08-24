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
	SetJSONFormat(jsonFormat)

	handler.SetReportTimestamp(true)

	return setLogger(handler)
}

func SetLevel(level string) {
	lvl, err := charmlog.ParseLevel(level)
	if err != nil {
		lvl = charmlog.InfoLevel
	}
	handler.SetLevel(lvl)
}

func SetJSONFormat(json bool) {
	if json {
		handler.SetFormatter(charmlog.JSONFormatter)
		setLogger(handler)
		return
	}
	handler.SetFormatter(charmlog.TextFormatter)
	setLogger(handler)
}

func setLogger(handle *charmlog.Logger) *slog.Logger {
	logger := slog.New(handle)
	slog.SetDefault(logger)
	return logger
}

// GetLogger returns the global logger instance.
// The logger should be configured via Setup before use.
func GetLogger() *slog.Logger {
	return slog.Default()
}
