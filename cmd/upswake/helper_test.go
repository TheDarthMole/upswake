package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ErrTimeout = errors.New("timeout")

func executeCommandWithContext(t *testing.T, cmd *cobra.Command, timeout time.Duration, args ...string) (output string, err error) {
	var buf bytes.Buffer
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	beforeSugar := sugar
	defer func() {
		sugar = beforeSugar
	}()

	beforeStderr := os.Stderr
	beforeStdout := os.Stdout
	defer func() {
		os.Stderr = beforeStderr
		os.Stdout = beforeStdout
	}()

	os.Stderr = w
	os.Stdout = w

	sugar = newMockLogger(w)

	cmd.SetOut(w)
	cmd.SetErr(w)
	cmd.SetArgs(args)
	os.Stderr = w

	// setup timeout for commands that can run indefinitely
	c := make(chan error, 1)
	go func() { c <- cmd.ExecuteContext(t.Context()) }()
	select {
	case err = <-c:
		// use err and reply
	case <-time.After(timeout):
		// set the error to be a timeout error
		err = ErrTimeout
		t.Context().Done()
	}

	w.Close()

	_, err1 := io.Copy(&buf, r)
	require.NoError(t, err1)

	return buf.String(), err
}

func newMockLogger(buf zapcore.WriteSyncer, options ...zap.Option) *zap.SugaredLogger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), buf, zap.DebugLevel)
	return zap.New(core).WithOptions(options...).Sugar()
}

func getStdoutStderr(t *testing.T, a func()) string {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	beforeStderr := os.Stderr
	beforeStdout := os.Stdout
	defer func() {
		os.Stderr = beforeStderr
		os.Stdout = beforeStdout
		w.Close()
		r.Close()
	}()

	os.Stderr = w
	os.Stdout = w

	a()

	var buf bytes.Buffer
	w.Close()
	_, err1 := io.Copy(&buf, r)
	require.NoError(t, err1)

	return buf.String()
}
