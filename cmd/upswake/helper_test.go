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

func NewTestLogger(pipeTo io.Writer) *zap.Logger {
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zap.CombineWriteSyncers(zapcore.AddSync(pipeTo)),
		zapcore.InfoLevel,
	))
}

func executeCommandWithContext(t *testing.T, cmdFunc func(_ *zap.SugaredLogger) *cobra.Command, timeout time.Duration, args ...string) (output string, err error) {
	var buf bytes.Buffer
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	beforeStderr := os.Stderr
	beforeStdout := os.Stdout
	defer func() {
		os.Stderr = beforeStderr
		os.Stdout = beforeStdout
	}()

	os.Stderr = w
	os.Stdout = w

	logger := NewTestLogger(w)
	sugar := logger.Sugar()

	cmd := cmdFunc(sugar)

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
