package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

var ErrTimeout = errors.New("timeout")

func newTestLoggerWithBuffer() (*slog.Logger, *bytes.Buffer) {
	logBuf := new(bytes.Buffer)
	handler := slog.NewJSONHandler(logBuf, nil)
	logger := slog.New(handler)
	return logger, logBuf
}

func newTestLogger() *slog.Logger {
	logger, _ := newTestLoggerWithBuffer()
	return logger
}

func executeCommandWithContext(t *testing.T, cmdFunc func(logger *slog.Logger) *cobra.Command, timeout time.Duration, args ...string) (output string, err error) {
	logBuf := new(bytes.Buffer)
	handler := slog.NewJSONHandler(logBuf, nil)
	logger := slog.New(handler)

	cmd := cmdFunc(logger)

	cmd.SetOut(logBuf)
	cmd.SetErr(logBuf)
	cmd.SetArgs(args)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()

	// setup timeout for commands that can run indefinitely
	c := make(chan error, 1)
	go func() { c <- cmd.ExecuteContext(ctx) }()
	select {
	case err = <-c:
		// use err and reply
	case <-time.After(timeout):
		// set the error to be a timeout error
		err = ErrTimeout
		cancel()
	}
	return logBuf.String(), err
}
