package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func executeCommandWithContext(t *testing.T, cmd *cobra.Command, args ...string) (output string, err error) {
	var buf bytes.Buffer
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	beforeSugar := sugar
	defer func() {
		sugar = beforeSugar
	}()

	sugar = newTestLogger(w)

	cmd.SetOut(w)
	cmd.SetErr(w)
	cmd.SetArgs(args)

	err = cmd.ExecuteContext(t.Context())

	w.Close()

	_, err1 := io.Copy(&buf, r)
	require.NoError(t, err1)

	return buf.String(), err
}

func newTestLogger(buf zapcore.WriteSyncer, options ...zap.Option) *zap.SugaredLogger {
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
