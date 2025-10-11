package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type syncWriter interface {
	Sync() error
	Write([]byte) (int, error)
	Read([]byte) (int, error)
}

func executeCommandWithContextC(t *testing.T, ctx context.Context, cmd *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	var buf bytes.Buffer
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()
	defer w.Close()

	//beforeSugar := sugar
	//defer func() {
	//	sugar = beforeSugar
	//}()

	sugar = newTestLogger(w)

	cmd.SetOut(w)
	cmd.SetErr(w)
	cmd.SetArgs(args)

	err = cmd.ExecuteContext(ctx)

	w.Close()

	io.Copy(&buf, r)

	return c, buf.String(), err
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
